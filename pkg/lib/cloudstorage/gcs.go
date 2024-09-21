package cloudstorage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

type GCSStorage struct {
	ServiceAccount  string
	Bucket          string
	CredentialsJSON []byte
	Clock           clock.Clock

	client *storage.Client
}

var _ Storage = &GCSStorage{}

func NewGCSStorage(
	credentialsJSON []byte,
	serviceAccount string,
	bucket string,
	c clock.Clock,
) (*GCSStorage, error) {
	// service account key is optional.
	// For backward compatibility, it is still unsupported.
	// When service account key is not provided, then the client is initialized with Application Default Credentials. (That is, without any option.ClientOption)
	// See https://pkg.go.dev/cloud.google.com/go/storage#hdr-Creating_a_Client
	var options []option.ClientOption
	if len(credentialsJSON) > 0 {
		options = append(options, option.WithCredentialsJSON(credentialsJSON))
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS: %w", err)
	}

	return &GCSStorage{
		ServiceAccount:  serviceAccount,
		Bucket:          bucket,
		CredentialsJSON: credentialsJSON,
		Clock:           c,
		client:          client,
	}, nil
}

func (s *GCSStorage) PresignPutObject(name string, header http.Header) (*http.Request, error) {
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
	opts := storage.SignedURLOptions{
		// We still need to tell the Client SDK which service account we want to sign the URL with.
		// https://pkg.go.dev/cloud.google.com/go/storage#hdr-Credential_requirements_for_signing
		GoogleAccessID: s.ServiceAccount,
		Method:         "PUT",
		Expires:        expires,
		ContentType:    header.Get("Content-Type"),
		Headers:        headers,
		MD5:            header.Get("Content-MD5"),
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := s.client.Bucket(s.Bucket).SignedURL(name, &opts)
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

func (s *GCSStorage) PresignGetOrHeadObject(name string, method string, expire time.Duration) (*url.URL, error) {
	now := s.Clock.NowUTC()
	expires := now.Add(expire)

	opts := storage.SignedURLOptions{
		// We still need to tell the Client SDK which service account we want to sign the URL with.
		// https://pkg.go.dev/cloud.google.com/go/storage#hdr-Credential_requirements_for_signing
		GoogleAccessID: s.ServiceAccount,
		Method:         method,
		Expires:        expires,
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := s.client.Bucket(s.Bucket).SignedURL(name, &opts)
	if err != nil {
		return nil, fmt.Errorf("failed to presign get or head request: %w", err)
	}

	u, _ := url.Parse(urlStr)

	return u, nil
}

func (s *GCSStorage) PresignHeadObject(name string, expire time.Duration) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "HEAD", expire)
}

func (s *GCSStorage) MakeDirector(extractKey func(r *http.Request) string) func(r *http.Request) {
	return func(r *http.Request) {
		key := extractKey(r)
		u, err := s.PresignGetOrHeadObject(key, "GET", PresignGetExpires)
		if err != nil {
			panic(err)
		}
		r.Host = ""
		r.URL = u
	}
}
