package cloudstorage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	raw "google.golang.org/api/storage/v1"

	"github.com/authgear/authgear-server/pkg/util/clock"
)

type GCSStorage struct {
	ServiceAccount  string
	Bucket          string
	CredentialsJSON []byte
	Clock           clock.Clock

	privateKey []byte
	client     *storage.Client
	service    *raw.Service
}

var _ Storage = &GCSStorage{}

func NewGCSStorage(
	credentialsJSON []byte,
	serviceAccount string,
	bucket string,
	c clock.Clock,
) (*GCSStorage, error) {
	s := &GCSStorage{
		ServiceAccount:  serviceAccount,
		Bucket:          bucket,
		CredentialsJSON: credentialsJSON,
		Clock:           c,
	}

	var j map[string]interface{}
	err := json.NewDecoder(bytes.NewReader(credentialsJSON)).Decode(&j)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials JSON: %w", err)
	}

	privateKeyStr, ok := j["private_key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing private")
	}
	s.privateKey = []byte(privateKeyStr)

	ctx := context.Background()
	service, err := raw.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS: %w", err)
	}
	s.service = service

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GCS: %w", err)
	}
	s.client = client

	return s, nil
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
		GoogleAccessID: s.ServiceAccount,
		PrivateKey:     s.privateKey,
		Method:         "PUT",
		Expires:        expires,
		ContentType:    header.Get("Content-Type"),
		Headers:        headers,
		MD5:            header.Get("Content-MD5"),
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := storage.SignedURL(s.Bucket, name, &opts)
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

func (s *GCSStorage) PresignGetOrHeadObject(name string, method string) (*url.URL, error) {
	now := s.Clock.NowUTC()
	expires := now.Add(PresignGetExpires)

	opts := storage.SignedURLOptions{
		GoogleAccessID: s.ServiceAccount,
		PrivateKey:     s.privateKey,
		Method:         method,
		Expires:        expires,
		Scheme:         storage.SigningSchemeV4,
	}
	urlStr, err := storage.SignedURL(s.Bucket, name, &opts)
	if err != nil {
		return nil, fmt.Errorf("failed to presign get or head request: %w", err)
	}

	u, _ := url.Parse(urlStr)

	return u, nil
}

func (s *GCSStorage) PresignHeadObject(name string) (*url.URL, error) {
	return s.PresignGetOrHeadObject(name, "HEAD")
}

func (s *GCSStorage) MakeDirector(extractKey func(r *http.Request) string) func(r *http.Request) {
	return func(r *http.Request) {
		key := extractKey(r)
		u, err := s.PresignGetOrHeadObject(key, "GET")
		if err != nil {
			panic(err)
		}
		r.Host = ""
		r.URL = u
	}
}
