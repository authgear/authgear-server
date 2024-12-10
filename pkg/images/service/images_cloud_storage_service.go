package service

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// PresignGetExpires is how long the presign GET request remains valid.
const PresignGetExpires time.Duration = 1 * duration.PerHour

// MaxContentLength is 10MiB.
const MaxContentLength = 10 * 1024 * 1024

type ImagesCloudStorageServiceStorage interface {
	PresignPutObject(ctx context.Context, name string, header http.Header) (*http.Request, error)
	PresignHeadObject(ctx context.Context, name string, expire time.Duration) (*url.URL, error)
	MakeDirector(extractKey func(r *http.Request) string, expire time.Duration) func(r *http.Request)
}

type ImagesCloudStorageServiceHTTPClient struct {
	*http.Client
}

func NewImagesCloudStorageServiceHTTPClient() ImagesCloudStorageServiceHTTPClient {
	return ImagesCloudStorageServiceHTTPClient{
		httputil.NewExternalClient(5 * time.Second),
	}
}

type ImagesCloudStorageService struct {
	HTTPClient ImagesCloudStorageServiceHTTPClient
	Storage    ImagesCloudStorageServiceStorage
}

func (p *ImagesCloudStorageService) PresignPutRequest(ctx context.Context, r *PresignUploadRequest) (*PresignUploadResponse, error) {
	r.Sanitize()

	contentLength := r.ContentLength()
	if contentLength <= 0 || contentLength > MaxContentLength {
		return nil, apierrors.NewBadRequest("asset too large")
	}

	key := r.Key

	err := p.checkDuplicate(ctx, key)
	if err != nil {
		return nil, err
	}

	httpHeader := r.HTTPHeader()
	httpRequest, err := p.Storage.PresignPutObject(ctx, key, httpHeader)
	if err != nil {
		return nil, err
	}

	resp := NewPresignUploadResponse(httpRequest, key)
	return &resp, nil
}

func (p *ImagesCloudStorageService) checkDuplicate(ctx context.Context, key string) error {
	u, err := p.Storage.PresignHeadObject(ctx, key, PresignGetExpires)
	if err != nil {
		return err
	}
	resp, err := httputil.HeadWithContext(ctx, p.HTTPClient.Client, u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil
	}
	return apierrors.AlreadyExists.WithReason("DuplicatedImage").Errorf("duplicated image")
}

func (p *ImagesCloudStorageService) MakeDirector(extractKey func(r *http.Request) string) func(r *http.Request) {
	return p.Storage.MakeDirector(extractKey, PresignGetExpires)
}
