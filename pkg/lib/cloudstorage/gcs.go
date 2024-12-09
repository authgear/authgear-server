package cloudstorage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

type GCSStorage struct {
	ServiceAccount  string
	Bucket          string
	CredentialsJSON []byte
	Clock           clock.Clock
}

var _ storage = &GCSStorage{}

func NewGCSStorage(
	credentialsJSON []byte,
	serviceAccount string,
	bucket string,
	c clock.Clock,
) (*GCSStorage, error) {
	return &GCSStorage{
		ServiceAccount:  serviceAccount,
		Bucket:          bucket,
		CredentialsJSON: credentialsJSON,
		Clock:           c,
	}, nil
}

func (s *GCSStorage) makeClient(ctx context.Context) (*gcs.Client, error) {
	// service account key is optional.
	// For backward compatibility, it is still unsupported.
	// When service account key is not provided, then the client is initialized with Application Default Credentials. (That is, without any option.ClientOption)
	// See https://pkg.go.dev/cloud.google.com/go/storage#hdr-Creating_a_Client
	var options []option.ClientOption
	if len(s.CredentialsJSON) > 0 {
		options = append(options, option.WithCredentialsJSON(s.CredentialsJSON))
	}
	client, err := gcs.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS: %w", err)
	}
	return client, nil
}

func (s *GCSStorage) PresignPutObject(ctx context.Context, name string, header http.Header) (*http.Request, error) {
	client, err := s.makeClient(ctx)
	if err != nil {
		return nil, err
	}

	now := s.Clock.NowUTC()

	// We must omit Content-type and Content-MD5 from header because they are special.
	var headers []string
	for name := range header {
		lower := strings.ToLower(name)
		if lower == "content-type" || lower == "content-md5" {
			continue
		}
		headers = append(headers, fmt.Sprintf("%s:%s", lower, header.Get(name)))
	}

	expires := now.Add(PresignPutExpires)
	opts := gcs.SignedURLOptions{
		// We still need to tell the Client SDK which service account we want to sign the URL with.
		// https://pkg.go.dev/cloud.google.com/go/storage#hdr-Credential_requirements_for_signing
		GoogleAccessID: s.ServiceAccount,
		Method:         "PUT",
		Expires:        expires,
		ContentType:    header.Get("Content-Type"),
		Headers:        headers,
		MD5:            header.Get("Content-MD5"),
		Scheme:         gcs.SigningSchemeV4,
	}
	urlStr, err := client.Bucket(s.Bucket).SignedURL(name, &opts)
	if err != nil {
		return nil, fmt.Errorf("failed to presign put request: %w", err)
	}

	u, _ := url.Parse(urlStr)
	req := http.Request{
		Method: "PUT",
		Header: header,
		URL:    u,
	}

	return &req, nil
}

func (s *GCSStorage) PresignGetOrHeadObject(ctx context.Context, name string, method string, expire time.Duration) (*url.URL, error) {
	client, err := s.makeClient(ctx)
	if err != nil {
		return nil, err
	}

	now := s.Clock.NowUTC()
	expires := now.Add(expire)

	opts := gcs.SignedURLOptions{
		// We still need to tell the Client SDK which service account we want to sign the URL with.
		// https://pkg.go.dev/cloud.google.com/go/storage#hdr-Credential_requirements_for_signing
		GoogleAccessID: s.ServiceAccount,
		Method:         method,
		Expires:        expires,
		Scheme:         gcs.SigningSchemeV4,
	}
	urlStr, err := client.Bucket(s.Bucket).SignedURL(name, &opts)
	if err != nil {
		return nil, fmt.Errorf("failed to presign get or head request: %w", err)
	}

	u, _ := url.Parse(urlStr)

	return u, nil
}

func (s *GCSStorage) PresignHeadObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	return s.PresignGetOrHeadObject(ctx, name, "HEAD", expire)
}

func (s *GCSStorage) PresignGetObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error) {
	return s.PresignGetOrHeadObject(ctx, name, "GET", expire)
}

func (s *GCSStorage) MakeDirector(extractKey func(r *http.Request) string, expire time.Duration) func(r *http.Request) {
	return func(r *http.Request) {
		key := extractKey(r)
		u, err := s.PresignGetOrHeadObject(r.Context(), key, "GET", expire)
		if err != nil {
			panic(err)
		}
		r.Host = ""
		r.URL = u
	}
}
