package service

import (
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

// PresignGetExpires is how long the presign GET request remains valid.
const PresignGetExpires time.Duration = 1 * duration.PerHour

// MaxContentLength is 10MiB.
const MaxContentLength = 10 * 1024 * 1024

type ImagesCloudStorageServiceStorage interface {
	PresignPutObject(name string, header http.Header) (*http.Request, error)
	PresignHeadObject(name string, expire time.Duration) (*url.URL, error)
	MakeDirector(extractKey func(r *http.Request) string, expire time.Duration) func(r *http.Request)
}

type ImagesCloudStorageService struct {
	Storage ImagesCloudStorageServiceStorage
}

func (p *ImagesCloudStorageService) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	r.Sanitize()

	contentLength := r.ContentLength()
	if contentLength <= 0 || contentLength > MaxContentLength {
		return nil, apierrors.NewBadRequest("asset too large")
	}

	key := r.Key

	err := p.checkDuplicate(key)
	if err != nil {
		return nil, err
	}

	httpHeader := r.HTTPHeader()
	httpRequest, err := p.Storage.PresignPutObject(key, httpHeader)
	if err != nil {
		return nil, err
	}

	resp := NewPresignUploadResponse(httpRequest, key)
	return &resp, nil
}

func (p *ImagesCloudStorageService) checkDuplicate(key string) error {
	u, err := p.Storage.PresignHeadObject(key, PresignGetExpires)
	if err != nil {
		return err
	}
	resp, err := http.Head(u.String())
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
