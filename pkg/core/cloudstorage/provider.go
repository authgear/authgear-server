package cloudstorage

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// ErrDuplicateAsset happens when the asset name is random and conflicts.
var ErrDuplicateAsset = skyerr.NewError(skyerr.Duplicated, "duplicate asset")

// Provider manipulates cloud storage.
type Provider interface {
	PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error)
	Sign(r *SignRequest) (*SignRequest, error)
	RewriteGetURL(u *url.URL, name string) (*url.URL, bool, error)
	AccessType(header http.Header) AccessType
	ProprietaryToStandard(header http.Header) http.Header
}
