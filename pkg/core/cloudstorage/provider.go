package cloudstorage

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// ErrDuplicateAsset happens when the asset name is random and conflicts.
var ErrDuplicateAsset = skyerr.AlreadyExists.WithReason("DuplicateAsset").New("duplicate asset")

// ErrAssetTooLarge happens when the asset exceeds MaxContentLength.
var ErrAssetTooLarge = skyerr.BadRequest.WithReason("AssetTooLarge").New("asset too large")

// MaxContentLength is 512MiB.
const MaxContentLength = 512 * 1024 * 1024

// Provider manipulates cloud storage.
type Provider interface {
	PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error)
	Sign(scheme string, host string, r *SignRequest) error
	Verify(r *http.Request) error
	PresignGetRequest(assetName string) (*url.URL, error)
	List(r *ListObjectsRequest) (*ListObjectsResponse, error)
	Delete(name string) error
	AccessType(header http.Header) AccessType
	ProprietaryToStandard(header http.Header) http.Header
}
