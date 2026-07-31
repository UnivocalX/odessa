package azure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func parse(location string) (container, blob string, err error) {
	u, err := url.Parse(location)
	if err != nil {
		return "", "", err
	}
	container = u.Host
	blob = strings.TrimPrefix(u.Path, "/")
	if container == "" {
		return "", "", fmt.Errorf("azure: missing container in %q", location)
	}
	return container, blob, nil
}

func isNotFound(err error) bool {
	var re *azcore.ResponseError
	return errors.As(err, &re) && re.StatusCode == 404
}

func (s *Store) Get(ctx context.Context, location string) (io.ReadCloser, error) {
	container, blob, err := parse(location)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.DownloadStream(ctx, container, blob, nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s *Store) Put(ctx context.Context, location string, r io.Reader) error {
	container, blob, err := parse(location)
	if err != nil {
		return err
	}
	_, err = s.client.UploadStream(ctx, container, blob, r, nil)
	return err
}

func (s *Store) Delete(ctx context.Context, location string) error {
	container, blob, err := parse(location)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteBlob(ctx, container, blob, nil)
	return err
}

func (s *Store) Available(ctx context.Context, location string) (bool, error) {
	container, blob, err := parse(location)
	if err != nil {
		return false, err
	}
	blobClient := s.client.ServiceClient().NewContainerClient(container).NewBlobClient(blob)
	_, err = blobClient.GetProperties(ctx, nil)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	container, blobPrefix, err := parse(prefix)
	if err != nil {
		return nil, err
	}
	var keys []string
	pager := s.client.NewListBlobsFlatPager(container, &azblob.ListBlobsFlatOptions{
		Prefix: &blobPrefix,
	})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range page.Segment.BlobItems {
			keys = append(keys, fmt.Sprintf("az://%s/%s", container, *item.Name))
		}
	}
	return keys, nil
}
