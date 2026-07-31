package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	"example.com/aether/internal/storage"
)

func init() {
	storage.Register("az", &Store{})
}

// Store is an Azure Blob Storage backend.
// Location format: az://container/blob-path
type Store struct {
	client *azblob.Client
}

// Option configures a Store.
type Option func(*Store) error

// WithConnectionString authenticates using an Azure storage connection string.
func WithConnectionString(connStr string) Option {
	return func(s *Store) error {
		client, err := azblob.NewClientFromConnectionString(connStr, nil)
		if err != nil {
			return err
		}
		s.client = client
		return nil
	}
}

// WithAccountKey authenticates with a storage account name and shared key.
func WithAccountKey(account, key string) Option {
	return func(s *Store) error {
		cred, err := azblob.NewSharedKeyCredential(account, key)
		if err != nil {
			return err
		}
		client, err := azblob.NewClientWithSharedKeyCredential(
			"https://"+account+".blob.core.windows.net/", cred, nil,
		)
		if err != nil {
			return err
		}
		s.client = client
		return nil
	}
}

// WithDefaultCredential uses the default Azure credential chain:
// environment variables, workload identity, managed identity, Azure CLI, etc.
func WithDefaultCredential(account string) Option {
	return func(s *Store) error {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return err
		}
		client, err := azblob.NewClient(
			"https://"+account+".blob.core.windows.net/", cred, nil,
		)
		if err != nil {
			return err
		}
		s.client = client
		return nil
	}
}

// Configure applies options to the globally-registered Azure store.
func Configure(opts ...Option) error {
	backend, ok := storage.Backend("az")
	if !ok {
		return fmt.Errorf("azure: backend not registered")
	}
	s := backend.(*Store)
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return err
		}
	}
	return nil
}
