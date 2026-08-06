package repository

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"gorm.io/gorm"
)

// SearchOption applies a filter or pagination scope to a blob search query.
type SearchOption func(*gorm.DB) *gorm.DB

func WithHashes(hashes ...string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(hashes) == 0 {
			return db
		}
		return db.Where("blobs.hash IN ?", hashes)
	}
}

func WithoutHashes(hashes ...string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(hashes) == 0 {
			return db
		}
		return db.Where("blobs.hash NOT IN ?", hashes)
	}
}

func WithMimeTypes(types ...string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(types) == 0 {
			return db
		}
		return db.Where("blobs.mime_type IN ?", types)
	}
}

func WithoutMimeTypes(types ...string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(types) == 0 {
			return db
		}
		return db.Where("blobs.mime_type NOT IN ?", types)
	}
}

func WithLabels(names ...string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(names) == 0 {
			return db
		}
		sub := db.Session(&gorm.Session{NewDB: true}).Model(&BlobLabel{}).
			Select("blob_labels.blob_id").
			Joins("JOIN labels ON labels.id = blob_labels.label_id").
			Where("labels.name IN ?", names).
			Group("blob_labels.blob_id").
			Having("COUNT(DISTINCT labels.name) = ?", len(names))
		return db.Where("blobs.id IN (?)", sub)
	}
}

func WithoutLabels(names ...string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(names) == 0 {
			return db
		}

		sub := db.Session(&gorm.Session{NewDB: true}).
			Model(&BlobLabel{}).
			Select("blob_labels.blob_id").
			Joins("JOIN labels ON labels.id = blob_labels.label_id").
			Where("labels.name IN ?", names).
			Group("blob_labels.blob_id").
			Having("COUNT(DISTINCT labels.name) = ?", len(names))

		return db.Where("blobs.id NOT IN (?)", sub)
	}
}

func WithLabelValues(kv map[string]string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(kv) == 0 {
			return db
		}
		sub := db.Session(&gorm.Session{NewDB: true}).Model(&BlobLabel{}).
			Select("blob_labels.blob_id")

		var conditions []string
		var args []any
		for name, value := range kv {
			conditions = append(conditions, "(labels.name = ? AND blob_labels.value = ?)")
			args = append(args, name, value)
		}
		orClause := ""
		for i, c := range conditions {
			if i > 0 {
				orClause += " OR "
			}
			orClause += c
		}
		sub = sub.Joins("JOIN labels ON labels.id = blob_labels.label_id").
			Where(orClause, args...).
			Group("blob_labels.blob_id").
			Having("COUNT(DISTINCT labels.name) = ?", len(kv))
		return db.Where("blobs.id IN (?)", sub)
	}
}

func WithoutLabelValues(kv map[string]string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(kv) == 0 {
			return db
		}

		sub := db.Session(&gorm.Session{NewDB: true}).
			Model(&BlobLabel{}).
			Select("blob_labels.blob_id")

		var conditions []string
		var args []any

		for name, value := range kv {
			conditions = append(conditions, "(labels.name = ? AND blob_labels.value = ?)")
			args = append(args, name, value)
		}

		orClause := strings.Join(conditions, " OR ")

		sub = sub.Joins("JOIN labels ON labels.id = blob_labels.label_id").
			Where(orClause, args...).
			Group("blob_labels.blob_id").
			Having("COUNT(DISTINCT labels.name) = ?", len(kv))

		return db.Where("blobs.id NOT IN (?)", sub)
	}
}

func WithURIPattern(pattern string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if pattern == "" {
			return db
		}
		sub := db.Session(&gorm.Session{NewDB: true}).
			Model(&Location{}).Select("blob_id").Where("uri LIKE ?", pattern)
		return db.Where("blobs.id IN (?)", sub)
	}
}

func WithoutURIPattern(pattern string) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if pattern == "" {
			return db
		}

		sub := db.Session(&gorm.Session{NewDB: true}).
			Model(&Location{}).
			Select("blob_id").
			Where("uri LIKE ?", pattern)

		return db.Where("blobs.id NOT IN (?)", sub)
	}
}

func WithMinSize(size int64) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("blobs.size >= ?", size)
	}
}

func WithMaxSize(size int64) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("blobs.size <= ?", size)
	}
}

func WithCursor(cursor uint) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if cursor == 0 {
			return db
		}
		return db.Where("blobs.id > ?", cursor)
	}
}

func WithLimit(limit int) SearchOption {
	return func(db *gorm.DB) *gorm.DB {
		if limit > 0 && limit <= 1000 {
			// Store as a GORM setting so SearchBlobs can read it.
			return db.Set("search:limit", limit)
		}
		return db
	}
}

// BlobSearchResult holds a single page of search results.
type BlobSearchResult struct {
	Blobs      []Blob `json:"blobs"`
	NextCursor uint   `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

// SearchBlobs returns a single page of blobs matching the given options.
// Use WithCursor and WithLimit to control pagination.
func (r *Repository) SearchBlobs(ctx context.Context, opts ...SearchOption) (*BlobSearchResult, error) {
	db := r.DB.WithContext(ctx).Model(&Blob{})
	for _, opt := range opts {
		db = opt(db)
	}

	limit := 100
	if v, ok := db.Get("search:limit"); ok {
		limit = v.(int)
	}

	var blobs []Blob
	if err := db.Order("blobs.id ASC").Limit(limit + 1).
		Preload("Locations").Preload("Labels").Preload("Labels.Label").
		Find(&blobs).Error; err != nil {
		return nil, fmt.Errorf("repository: search blobs: %w", err)
	}

	result := &BlobSearchResult{}
	if len(blobs) > limit {
		result.HasMore = true
		blobs = blobs[:limit]
	}
	result.Blobs = blobs
	if len(blobs) > 0 {
		result.NextCursor = blobs[len(blobs)-1].ID
	}
	return result, nil
}

// SearchBlobsGenarator returns an iterator that yields pages of blobs matching
// the given search options, automatically paging through all results.
func (r *Repository) SearchBlobsGenarator(ctx context.Context, opts ...SearchOption) iter.Seq2[[]Blob, error] {
	return func(yield func([]Blob, error) bool) {
		cursor := uint(0)
		for {
			pageOpts := make([]SearchOption, len(opts))
			copy(pageOpts, opts)
			if cursor > 0 {
				pageOpts = append(pageOpts, WithCursor(cursor))
			}

			result, err := r.SearchBlobs(ctx, pageOpts...)
			if err != nil {
				yield(nil, err)
				return
			}

			if len(result.Blobs) == 0 {
				return
			}

			if !yield(result.Blobs, nil) {
				return
			}

			if !result.HasMore {
				return
			}
			cursor = result.NextCursor
		}
	}
}
