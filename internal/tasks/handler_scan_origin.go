package tasks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/storage"
	"github.com/UnivocalX/universe"
)

const TaskTypeScanOrigin = "scan_origin"

// ScanOriginHandler owns the full lifecycle of scan-origin tasks:
// claiming pending scans, running the file-discovery pipeline, persisting
// blobs, and applying label rules.
type ScanOriginHandler struct {
	repo *repository.Repository
	reg  *storage.Registry
}

// ScanResults is persisted as the scan-origin task's JSON results payload.
type ScanResults struct {
	Created int      `json:"created"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// fileInfo holds the metadata produced by the scan pipeline for a single file.
type fileInfo struct {
	URI      string
	Hash     string
	MimeType string
	Size     int64
}

// hashEntry groups files that share the same content hash together with
// the label assignments that should be applied to them.
type hashEntry struct {
	files       []fileInfo
	assignments []repository.LabelAssignment
}

// NewScanOriginHandler creates a handler wired to the given repository and
// storage registry.
func NewScanOriginHandler(repo *repository.Repository, reg *storage.Registry) *ScanOriginHandler {
	return &ScanOriginHandler{repo: repo, reg: reg}
}

// Type returns the task type identifier used for registration.
func (h *ScanOriginHandler) Type() string {
	return TaskTypeScanOrigin
}

// Claim atomically claims up to `limit` pending scan-origin jobs.
func (h *ScanOriginHandler) Claim(ctx context.Context, limit int) ([]Job, error) {
	scans, err := h.repo.ClaimScanOrigins(ctx, limit)
	if err != nil {
		return nil, err
	}

	jobs := make([]Job, len(scans))
	for i, scan := range scans {
		jobs[i] = Job{
			ID:      scan.ID,
			Type:    TaskTypeScanOrigin,
			Attempt: scan.Attempts,
			Payload: scan,
		}
	}
	return jobs, nil
}

// Complete marks a scan-origin job as successfully completed.
func (h *ScanOriginHandler) Complete(ctx context.Context, id uint, results any) error {
	return h.repo.CompleteScanOrigin(ctx, id, results)
}

// Fail marks a scan-origin job as failed and re-enqueues it if under max attempts.
func (h *ScanOriginHandler) Fail(ctx context.Context, id uint, maxAttempts int) error {
	return h.repo.FailScanOrigin(ctx, id, maxAttempts)
}

// Handle is the main entry point for processing a single scan-origin job.
// It runs the file-discovery pipeline, persists discovered blobs, and
// applies any label rules attached to the scan request.
func (h *ScanOriginHandler) Handle(ctx context.Context, job Job) (any, error) {
	log := slog.With("task", job.Type, "job_id", job.ID, "attempt", job.Attempt+1)

	scan, ok := job.Payload.(repository.ScanOrigin)
	if !ok {
		return nil, fmt.Errorf("unexpected payload type: %T", job.Payload)
	}

	// Create a cancellable processing context and start a watcher that polls
	// the scan record; if a cancel request is observed, cancel processing.
	procCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Prefer event-based notifications: subscribe to NOTIFY/LISTEN for
	// scan cancellations. Fall back to the polling watcher if subscribe fails.
	ch, stop, err := h.repo.SubscribeScanCancels(procCtx)
	if err != nil {
		doneCh := h.startCancelWatcher(scan.ID, cancel)
		defer close(doneCh)
	} else {
		defer stop()
		// react to notifications for our scan id
		go func() {
			for id := range ch {
				if id == scan.ID {
					slog.Info("scan cancelled via NOTIFY", "scan_id", scan.ID)
					cancel()
					return
				}
			}
		}()
	}

	origin, err := h.repo.GetOrigin(procCtx, scan.OriginID)
	if err != nil {
		return nil, fmt.Errorf("get origin %d: %w", scan.OriginID, err)
	}

	// Run the scan pipeline: list, hash, and detect MIME types.
	files, scanErr := ScanOriginPipeline(procCtx, h.reg, string(origin.URI))
	results := &ScanResults{}
	collectScanErrors(results, scanErr)

	if procCtx.Err() != nil {
		log.Info("aborting scan after cancellation", "scan_id", scan.ID)
		return nil, nil
	}

	// Upsert blobs and their storage locations.
	if err := h.persistBlobs(procCtx, files, results); err != nil {
		return nil, err
	}

	if procCtx.Err() != nil {
		log.Info("aborting scan after cancellation (persist)", "scan_id", scan.ID)
		return nil, nil
	}

	// Apply label rules (pattern → label assignments).
	if err := h.applyRules(procCtx, origin, scan, files, results); err != nil {
		return nil, err
	}

	if procCtx.Err() != nil {
		log.Info("aborting scan after cancellation (apply rules)", "scan_id", scan.ID)
		return nil, nil
	}

	if results.Created == 0 {
		return results, fmt.Errorf("no blobs created: %d files scanned, %d failed", len(files), results.Failed)
	}

	log.Info("scan complete", "created", results.Created, "failed", results.Failed)
	return results, nil
}

// startCancelWatcher begins a background goroutine that checks the database
// for a cancelled scan record. It returns a channel which can be closed to
// stop the watcher. When the scan is found cancelled the provided cancel
// function is invoked.
func (h *ScanOriginHandler) startCancelWatcher(scanID uint, cancel context.CancelFunc) chan struct{} {
	doneCh := make(chan struct{})

	// Immediate check to avoid waiting for the first tick.
	if s, err := h.repo.GetScanOrigin(context.Background(), scanID); err == nil {
		if s.Status == repository.StatusCancelled {
			slog.Info("scan cancelled before processing (watcher immediate)", "scan_id", scanID)
			cancel()
			return doneCh
		}
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-doneCh:
				return
			case <-ticker.C:
				s, err := h.repo.GetScanOrigin(context.Background(), scanID)
				if err != nil {
					continue
				}
				if s.Status == repository.StatusCancelled {
					slog.Info("scan cancelled during processing", "scan_id", scanID)
					cancel()
					return
				}
			}
		}
	}()
	return doneCh
}

// collectScanErrors extracts per-file errors from the pipeline's MultiError
// and records them in the results. Each error is logged individually so
// failures are visible in the worker's log output.
func collectScanErrors(results *ScanResults, err error) {
	var scanErr *MultiError
	if !errors.As(err, &scanErr) {
		return
	}
	results.Failed += len(scanErr.Errs)
	for uri, e := range scanErr.Errs {
		msg := fmt.Sprintf("%s: %s", uri, e)
		results.Errors = append(results.Errors, msg)
		slog.Warn("scan file error", "uri", uri, "error", e)
	}
}

// persistBlobs upserts the discovered files as blobs with their storage
// locations. Errors from individual upserts are non-fatal and accumulated
// in results.
func (h *ScanOriginHandler) persistBlobs(ctx context.Context, files []fileInfo, results *ScanResults) error {
	if len(files) == 0 {
		return nil
	}
	batch, err := h.repo.BatchCreateBlobs(ctx, filesToBlobInputs(files))
	if err != nil {
		return fmt.Errorf("batch create blobs: %w", err)
	}
	results.Created += batch.Created
	results.Failed += batch.Failed
	for _, e := range batch.Errors {
		results.Errors = append(results.Errors, e)
		slog.Warn("persist blob error", "error", e)
	}
	return nil
}

// applyRules resolves the effective label rules by merging the origin's
// default rules with any scan-level overrides (scan rules take precedence
// for the same glob pattern), then delegates to applyLabelRules.
func (h *ScanOriginHandler) applyRules(ctx context.Context, origin *repository.Origin, scan repository.ScanOrigin, files []fileInfo, results *ScanResults) error {
	// Start with the origin's default rules.
	var rules repository.LabelRules
	if len(origin.Rules) > 0 {
		if err := json.Unmarshal(origin.Rules, &rules); err != nil {
			return fmt.Errorf("parse origin label rules: %w", err)
		}
	}

	// Merge scan-level overrides: same pattern key replaces origin's.
	if len(scan.Rules) > 0 {
		var scanRules repository.LabelRules
		if err := json.Unmarshal(scan.Rules, &scanRules); err != nil {
			return fmt.Errorf("parse scan label rules: %w", err)
		}
		if rules == nil {
			rules = scanRules
		} else {
			for pattern, assignments := range scanRules {
				rules[pattern] = assignments
			}
		}
	}

	if len(rules) == 0 {
		return nil
	}
	labelErrs := h.applyLabelRules(ctx, files, rules)
	results.Failed += len(labelErrs)
	results.Errors = append(results.Errors, labelErrs...)
	return nil
}

// ScanOriginPipeline resolves the storage backend, lists all files under the
// URI, and hashes them concurrently. Returns successful results and a
// *MultiError if any individual files failed (use errors.As to extract).
func ScanOriginPipeline(ctx context.Context, reg *storage.Registry, uri string) ([]fileInfo, error) {
	store, err := reg.Resolve(uri)
	if err != nil {
		return nil, fmt.Errorf("resolve backend for %q: %w", uri, err)
	}

	paths, err := store.List(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("list files at %q: %w", uri, err)
	}

	slog.Info("scanning origin", "uri", uri, "files", len(paths))

	stream := universe.Map(
		universe.From(ctx, paths...),
		func(path string) (fileInfo, error) {
			reader, err := store.Get(ctx, path)
			if err != nil {
				return fileInfo{URI: path}, fmt.Errorf("get %q: %w", path, err)
			}
			hash, mimeType, size, err := processFile(reader)
			if err != nil {
				return fileInfo{URI: path}, fmt.Errorf("process %q: %w", path, err)
			}
			return fileInfo{URI: path, Hash: hash, MimeType: mimeType, Size: size}, nil
		},
	).Buffer(16).Concurrent(4).Execute()

	var results []fileInfo
	errs := make(map[string]error)
	stream.ForEach(ctx, func(fi fileInfo, err error) {
		if err != nil {
			errs[fi.URI] = err
			return
		}
		results = append(results, fi)
	})

	if len(errs) > 0 {
		return results, &MultiError{Errs: errs}
	}
	return results, nil
}

// processFile reads the stream once, computing the SHA-256 hash, MIME type,
// and byte size in a single pass.
func processFile(r io.ReadCloser) (hash string, mimeType string, size int64, err error) {
	defer r.Close()

	h := sha256.New()

	// Read first 512 bytes for MIME detection.
	buf := make([]byte, 512)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", "", 0, err
	}
	buf = buf[:n]
	mimeType = http.DetectContentType(buf)

	h.Write(buf)
	size = int64(n)

	// Stream the rest through the hasher, counting bytes.
	written, err := io.Copy(h, r)
	if err != nil {
		return "", "", 0, err
	}
	size += written

	hash = hex.EncodeToString(h.Sum(nil))
	return hash, mimeType, size, nil
}

// applyLabelRules orchestrates the label-assignment pipeline:
//  1. Resolve label names → database IDs.
//  2. Group files by content hash and match them against glob rules.
//  3. Iterate through blobs via SearchBlobs (paginated generator).
//  4. For each page, build and execute the batch label assignment immediately
//     so we never hold millions of blob records in memory.
//
// All errors are non-fatal, logged, and returned as string slices.
func (h *ScanOriginHandler) applyLabelRules(ctx context.Context, files []fileInfo, rules repository.LabelRules) []string {
	var errs []string

	labelIDs, resolveErrs := h.resolveLabelIDs(ctx, rules)
	errs = append(errs, resolveErrs...)

	hashMap, allHashes := groupFilesByHash(files, rules)
	if len(allHashes) == 0 {
		return errs
	}

	// Process hashes in chunks of 1000 to avoid oversized IN clauses.
	const chunkSize = 1000
	for i := 0; i < len(allHashes); i += chunkSize {
		end := i + chunkSize
		if end > len(allHashes) {
			end = len(allHashes)
		}
		chunk := allHashes[i:end]

		// Iterate pages from the generator — each page is processed and
		// discarded before the next is fetched, keeping memory flat.
		for blobs, err := range h.repo.SearchBlobs(ctx,
			repository.WithHashes(chunk...),
			repository.WithLimit(chunkSize),
		) {
			if err != nil {
				msg := fmt.Sprintf("search blobs: %s", err)
				slog.Error("search blobs failed", "error", err, "chunk_start", i)
				errs = append(errs, msg)
				break
			}

			// Build label inputs for this page of blobs.
			var inputs []repository.BatchLabelInput
			for idx := range blobs {
				blob := &blobs[idx]
				entry, ok := hashMap[blob.Hash]
				if !ok {
					continue
				}
				for _, a := range entry.assignments {
					lid, ok := labelIDs[a.Label]
					if !ok {
						continue
					}
					inputs = append(inputs, repository.BatchLabelInput{
						BlobID:  blob.ID,
						LabelID: lid,
						Value:   a.Value,
					})
				}
				// Mark hash as seen so we can detect missing ones later.
				delete(hashMap, blob.Hash)
			}

			if len(inputs) == 0 {
				continue
			}

			result, err := h.repo.BatchAssignLabels(ctx, inputs)
			if err != nil {
				msg := fmt.Sprintf("batch assign labels: %s", err)
				slog.Error("batch assign labels failed", "error", err)
				errs = append(errs, msg)
				continue
			}
			for _, e := range result.Errors {
				slog.Warn("label assignment error", "error", e)
			}
			errs = append(errs, result.Errors...)
		}
	}

	// Report any hashes that were never found.
	for hash, entry := range hashMap {
		uris := make([]string, len(entry.files))
		for i, fi := range entry.files {
			uris[i] = fi.URI
		}
		errs = append(errs, fmt.Sprintf("blob not found for hash %s, affected files: %v", hash, uris))
	}

	return errs
}

// resolveLabelIDs looks up each unique label name referenced in the rules
// and returns a name→ID map. Unknown labels are logged and skipped.
func (h *ScanOriginHandler) resolveLabelIDs(ctx context.Context, rules repository.LabelRules) (map[string]uint, []string) {
	var errs []string
	labelIDs := make(map[string]uint)
	for _, assignments := range rules {
		for _, a := range assignments {
			if _, ok := labelIDs[a.Label]; ok {
				continue
			}
			label, err := h.repo.GetLabelByName(ctx, a.Label)
			if err != nil {
				msg := fmt.Sprintf("label %q not found: %s", a.Label, err)
				slog.Warn("resolve label failed", "label", a.Label, "error", err)
				errs = append(errs, msg)
				continue
			}
			labelIDs[a.Label] = label.ID
		}
	}
	return labelIDs, errs
}

// groupFilesByHash groups files by their content hash and collects the label
// assignments that match each file's URI against the glob rules.
// Returns the hash→entry map and a deduplicated slice of all hashes.
func groupFilesByHash(files []fileInfo, rules repository.LabelRules) (map[string]*hashEntry, []string) {
	hashMap := make(map[string]*hashEntry)
	var allHashes []string

	for _, fi := range files {
		assignments := matchingAssignments(fi.URI, rules)
		if len(assignments) == 0 {
			continue
		}
		entry, ok := hashMap[fi.Hash]
		if !ok {
			entry = &hashEntry{}
			hashMap[fi.Hash] = entry
			allHashes = append(allHashes, fi.Hash)
		}
		entry.files = append(entry.files, fi)
		entry.assignments = append(entry.assignments, assignments...)
	}
	return hashMap, allHashes
}

// matchingAssignments returns label assignments whose glob pattern matches
// the file URI's basename. The wildcard "*" matches all files.
func matchingAssignments(uri string, rules repository.LabelRules) []repository.LabelAssignment {
	var out []repository.LabelAssignment
	base := filepath.Base(uri)

	for pattern, labels := range rules {
		if pattern == "*" {
			out = append(out, labels...)
			continue
		}
		if matched, _ := filepath.Match(pattern, base); matched {
			out = append(out, labels...)
		}
	}
	return out
}

// filesToBlobInputs converts pipeline fileInfo results into the batch-create
// input format expected by the repository.
func filesToBlobInputs(files []fileInfo) []repository.BlobBatchCreateInput {
	inputs := make([]repository.BlobBatchCreateInput, len(files))
	for i, f := range files {
		inputs[i] = repository.BlobBatchCreateInput{
			Hash:     f.Hash,
			URI:      f.URI,
			MimeType: f.MimeType,
			Size:     f.Size,
		}
	}
	return inputs
}
