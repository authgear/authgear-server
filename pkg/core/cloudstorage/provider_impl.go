package cloudstorage

import (
	"fmt"
	"net/http"
)

type providerImpl struct {
	storage Storage
	appID   string
}

var _ Provider = &providerImpl{}

func NewProvider(appID string, storage Storage) Provider {
	return &providerImpl{
		appID:   appID,
		storage: storage,
	}
}

func (p *providerImpl) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	assetName, err := r.DeriveAssetName()
	if err != nil {
		return nil, err
	}

	r.SetCacheControl()

	// This is the final name in the storage.
	// It must not start with a leading slash because
	// /a/b is treated as <empty> / a / b by Azure Storage.
	assetID := fmt.Sprintf("%s/%s", p.appID, assetName)

	r.RemoveEmptyHeaders()

	// Check duplicatae if the name is random.
	if !r.IsCustomName() {
		err = p.checkDuplicate(assetID)
		if err != nil {
			return nil, err
		}
	}

	httpHeader := r.HTTPHeader()
	httpRequest, err := p.storage.PresignPutObject(assetID, r.Access, httpHeader)
	if err != nil {
		return nil, err
	}

	resp := NewPresignUploadResponse(httpRequest, assetName)
	return &resp, nil
}

func (p *providerImpl) checkDuplicate(assetID string) error {
	u, err := p.storage.PresignHeadObject(assetID)
	if err != nil {
		return err
	}
	resp, err := http.Head(u.String())
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		return nil
	}
	return ErrDuplicateAsset
}

func (p *providerImpl) Sign(r *SignRequest) (*SignRequest, error) {
	for i, assetItem := range r.Assets {
		u, err := p.storage.PresignGetObject(assetItem.AssetID)
		if err != nil {
			return nil, err
		}
		r.Assets[i].URL = u.String()
	}
	return r, nil
}
