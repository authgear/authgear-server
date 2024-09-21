package cloudstorage

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

// MaxContentLength is 10MiB.
const MaxContentLength = 10 * 1024 * 1024

type Provider struct {
	Storage Storage
}

func (p *Provider) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
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

func (p *Provider) checkDuplicate(key string) error {
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

func (p *Provider) MakeDirector(extractKey func(r *http.Request) string) func(r *http.Request) {
	return p.Storage.MakeDirector(extractKey)
}
