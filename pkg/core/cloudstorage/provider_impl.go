package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
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

func (p *providerImpl) GetAssetID(assetName string) string {
	// This is the final name in the storage.
	// It must not start with a leading slash because
	// /a/b is treated as <empty> / a / b by Azure Storage.
	assetID := fmt.Sprintf("%s/%s", p.appID, assetName)
	return assetID
}

func (p *providerImpl) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	assetName, err := r.DeriveAssetName()
	if err != nil {
		return nil, err
	}

	r.SetCacheControl()

	r.RemoveEmptyHeaders()

	assetID := p.GetAssetID(assetName)

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
		assetID := p.GetAssetID(assetItem.AssetName)
		u, err := p.storage.PresignGetObject(assetID)
		if err != nil {
			return nil, err
		}
		r.Assets[i].URL = u.String()
	}
	return r, nil
}

func (p *providerImpl) RewriteGetURL(u *url.URL, name string) (*url.URL, bool, error) {
	assetID := p.GetAssetID(name)
	return p.storage.RewriteGetURL(u, assetID)
}

func (p *providerImpl) ProprietaryToStandard(header http.Header) http.Header {
	return p.storage.ProprietaryToStandard(header)
}

func (p *providerImpl) AccessType(header http.Header) AccessType {
	return p.storage.AccessType(header)
}
