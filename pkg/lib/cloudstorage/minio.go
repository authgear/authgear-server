package cloudstorage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string

	client *minio.Client
}

var _ storage = (*MinIOStorage)(nil)

// MinIOStorage takes endpoint, accessKeyID, and secretAccessKey to construct a minio.Client under the hood.
// Contradictory to minio.Client, endpoint MUST BE a http URL or https URL.
func NewMinIOStorage(endpoint string, bucketName string, accessKeyID string, secretAccessKey string) (*MinIOStorage, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var secure bool
	switch u.Scheme {
	case "http":
		secure = false
	case "https":
		secure = true
	default:
		return nil, fmt.Errorf("minio: endpoint MUST BE a http URL or https URL: %v", endpoint)
	}

	var emptyToken string
	options := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, emptyToken),
		Secure: secure,
	}
	client, err := minio.New(u.Host, options)
	if err != nil {
		return nil, err
	}

	return &MinIOStorage{
		Endpoint:        endpoint,
		Bucket:          bucketName,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		client:          client,
	}, nil
}

func (s *MinIOStorage) PresignPutObject(ctx context.Context, name string, header http.Header) (*http.Request, error) {
	params := url.Values{}
	u, err := s.client.PresignHeader(ctx, "PUT", s.Bucket, name, PresignPutExpires, params, header)
	if err != nil {
		return nil, fmt.Errorf("failed to presign put request: %w", err)
	}

	return &http.Request{
		Method: "PUT",
		URL:    u,
		Header: header,
	}, nil
}

func (s *MinIOStorage) PresignHeadObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	params := url.Values{}
	u, err := s.client.PresignedHeadObject(ctx, s.Bucket, name, expire, params)
	if err != nil {
		return nil, fmt.Errorf("failed to presign head request: %w", err)
	}

	return u, nil
}

func (s *MinIOStorage) PresignGetObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	params := url.Values{}
	u, err := s.client.PresignedGetObject(ctx, s.Bucket, name, expire, params)
	if err != nil {
		return nil, fmt.Errorf("failed to presign get request: %w", err)
	}

	return u, nil
}

func (s *MinIOStorage) MakeDirector(extractKey func(r *http.Request) string, expire time.Duration) func(r *http.Request) {
	return func(r *http.Request) {
		key := extractKey(r)
		params := url.Values{}
		u, err := s.client.PresignedGetObject(r.Context(), s.Bucket, key, expire, params)
		if err != nil {
			panic(fmt.Errorf("failed to presign get request: %w", err))
		}

		r.Host = ""
		r.URL = u
	}
}
