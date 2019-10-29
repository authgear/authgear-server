package cloudstorage

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// ErrDuplicateAsset happens when the asset name is random and conflicts.
var ErrDuplicateAsset = skyerr.NewError(skyerr.Duplicated, "duplicate asset")

// ErrTooLargeAsset happens when the asset exceeds MaxContentLength.
var ErrTooLargeAsset = skyerr.NewError(skyerr.BadRequest, "too large asset")

// MaxContentLength is 512MiB.
const MaxContentLength = 512 * 1024 * 1024

// Provider manipulates cloud storage.
type Provider interface {
	PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error)
	Sign(scheme string, host string, r *SignRequest) (*SignRequest, error)
	RewriteGetURL(u *url.URL, name string) (*url.URL, bool, error)
	List(r *ListObjectsRequest) (*ListObjectsResponse, error)
	Delete(name string) error
	AccessType(header http.Header) AccessType
	ProprietaryToStandard(header http.Header) http.Header
}
