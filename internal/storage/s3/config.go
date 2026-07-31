package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/UnivocalX/odessa/internal/storage"
)

func init() {
	storage.Register("s3", &Store{})
}

// Store is an AWS S3 (or S3-compatible) storage backend.
// Location format: s3://bucket/key
type Store struct {
	client *awss3.Client
}

// Option configures a Store.
type Option func(*Store) error

// WithConfig uses the provided AWS config directly.
func WithConfig(cfg aws.Config) Option {
	return func(s *Store) error {
		s.client = awss3.NewFromConfig(cfg)
		return nil
	}
}

// WithRegion loads credentials from the default chain and sets the region.
func WithRegion(region string) Option {
	return func(s *Store) error {
		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
		if err != nil {
			return err
		}
		s.client = awss3.NewFromConfig(cfg)
		return nil
	}
}

// WithEndpoint targets a custom S3-compatible endpoint (e.g. MinIO).
// Path-style addressing is enabled automatically.
func WithEndpoint(endpoint, region string) Option {
	return func(s *Store) error {
		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
		if err != nil {
			return err
		}
		s.client = awss3.NewFromConfig(cfg, func(o *awss3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
		return nil
	}
}

// Configure applies options to the globally-registered S3 store.
func Configure(opts ...Option) error {
	backend, ok := storage.Backend("s3")
	if !ok {
		return fmt.Errorf("s3: backend not registered")
	}
	s := backend.(*Store)
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return err
		}
	}
	if s.client == nil {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return fmt.Errorf("s3: load default config: %w", err)
		}
		s.client = awss3.NewFromConfig(cfg)
	}
	return nil
}
