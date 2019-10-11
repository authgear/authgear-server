package cloudstorage

import (
	"errors"
)

// ErrDuplicateAsset happens when the asset name is random and conflicts.
var ErrDuplicateAsset = errors.New("duplicate asset")

// Provider manipulates cloud storage.
type Provider interface {
	PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error)
}
