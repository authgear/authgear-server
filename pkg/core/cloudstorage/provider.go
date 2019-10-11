package cloudstorage

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

// ErrDuplicateAsset happens when the asset name is random and conflicts.
var ErrDuplicateAsset = skyerr.NewError(skyerr.Duplicated, "duplicate asset")

// Provider manipulates cloud storage.
type Provider interface {
	PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error)
}
