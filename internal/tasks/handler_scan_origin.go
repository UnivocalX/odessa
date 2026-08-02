package tasks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/storage"
	"github.com/UnivocalX/universe"
)

const TaskTypeScanOrigin = "scan_origin"

// ScanOriginHandler owns the full lifecycle of scan origin tasks.
type ScanOriginHandler struct {
	repo *repository.Repository
	reg  *storage.Registry
}

func NewScanOriginHandler(repo *repository.Repository, reg *storage.Registry) *ScanOriginHandler {
	return &ScanOriginHandler{repo: repo, reg: reg}
}

func (h *ScanOriginHandler) Type() string {
	return TaskTypeScanOrigin
}

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

func (h *ScanOriginHandler) Complete(ctx context.Context, id uint, results any) error {
	return h.repo.CompleteScanOrigin(ctx, id, results)
}

func (h *ScanOriginHandler) Fail(ctx context.Context, id uint, maxAttempts int) error {
	return h.repo.FailScanOrigin(ctx, id, maxAttempts)
}

func (h *ScanOriginHandler) Handle(ctx context.Context, job Job) (any, error) {
	log := slog.With("task", job.Type, "job_id", job.ID, "attempt", job.Attempt+1)

	// validate payload as scan origin record
	scan, ok := job.Payload.(repository.ScanOrigin)
	if !ok {
		return nil, fmt.Errorf("unexpected payload type: %T", job.Payload)
	}

	// check origin exist
	origin, err := h.repo.GetOrigin(ctx, scan.OriginID)
	if err != nil {
		return nil, fmt.Errorf("get origin %d: %w", scan.OriginID, err)
	}

	// hash files
	files, scanErr := ScanOriginPipeline(ctx, h.reg, string(origin.URI))

	// Collect successes and errors.
	results := &ScanResults{}
	var inputs []repository.BlobBatchCreateInput

	if scanErr != nil {
		return nil, scanErr
	}

	for _, f := range files {
		if f.Err != nil {
			results.Failed++
			results.Errors = append(results.Errors, f.Err.Error())
			continue
		}
		inputs = append(inputs, repository.BlobBatchCreateInput{
			Hash: f.Hash,
			URI:  f.URI,
		})
	}

	// Batch create blobs + locations.
	if len(inputs) > 0 {
		batch, err := h.repo.BatchCreateBlobs(ctx, inputs)
		if err != nil {
			return nil, fmt.Errorf("batch create blobs: %w", err)
		}
		results.Created += batch.Created
		results.Failed += batch.Failed
		results.Errors = append(results.Errors, batch.Errors...)
	}

	if results.Created == 0 {
		return results, fmt.Errorf("no blobs created: %d files scanned, %d failed", len(files), results.Failed)
	}

	log.Info("scan complete", "created", results.Created, "failed", results.Failed)
	return results, nil
}

// ScanResults is persisted as the scan's JSON results.
type ScanResults struct {
	Created int      `json:"created"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

type fileInfo struct {
	URI  string
	Hash string
	Err  error
}

// ScanOriginPipeline resolves the storage backend, lists all files under the
// URI, and hashes them concurrently. Returns a slice of fileInfo with per-file
// results (Hash on success, Err on failure).
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

			hash, err := hashStream(reader)
			if err != nil {
				return fileInfo{URI: path}, fmt.Errorf("hash %q: %w", path, err)
			}

			return fileInfo{URI: path, Hash: hash}, nil
		},
	).Buffer(16).Concurrent(4).Execute()

	var results []fileInfo
	stream.ForEach(ctx, func(fi fileInfo, err error) {
		fi.Err = err
		results = append(results, fi)
	})

	return results, nil
}

func hashStream(r io.ReadCloser) (string, error) {
	h := sha256.New()
	defer r.Close()

	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
