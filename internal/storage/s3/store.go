package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func parse(location string) (bucket, key string, err error) {
	u, err := url.Parse(location)
	if err != nil {
		return "", "", err
	}
	bucket = u.Host
	key = strings.TrimPrefix(u.Path, "/")
	if bucket == "" {
		return "", "", fmt.Errorf("s3: missing bucket in %q", location)
	}
	return bucket, key, nil
}

func isNotFound(err error) bool {
	var re *smithyhttp.ResponseError
	return errors.As(err, &re) && re.HTTPStatusCode() == 404
}

func (s *Store) Get(ctx context.Context, location string) (io.ReadCloser, error) {
	bucket, key, err := parse(location)
	if err != nil {
		return nil, err
	}
	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *Store) Put(ctx context.Context, location string, r io.Reader) error {
	bucket, key, err := parse(location)
	if err != nil {
		return err
	}
	_, err = s.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	return err
}

func (s *Store) Delete(ctx context.Context, location string) error {
	bucket, key, err := parse(location)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *Store) Available(ctx context.Context, location string) (bool, error) {
	bucket, key, err := parse(location)
	if err != nil {
		return false, err
	}
	_, err = s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) List(ctx context.Context, prefix string) ([]string, error) {
	bucket, keyPrefix, err := parse(prefix)
	if err != nil {
		return nil, err
	}
	var keys []string
	paginator := awss3.NewListObjectsV2Paginator(s.client, &awss3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(keyPrefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			keys = append(keys, fmt.Sprintf("s3://%s/%s", bucket, aws.ToString(obj.Key)))
		}
	}
	return keys, nil
}
