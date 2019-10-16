package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

func (p *providerImpl) AssetNameToAssetID(assetName string) string {
	// This is the final name in the storage.
	// It must not start with a leading slash because
	// /a/b is treated as <empty> / a / b by Azure Storage.
	return fmt.Sprintf("%s/%s", p.appID, assetName)
}

func (p *providerImpl) AssetIDToAssetName(assetID string) string {
	return strings.TrimPrefix(assetID, fmt.Sprintf("%s/", p.appID))
}

func (p *providerImpl) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	assetName, err := r.DeriveAssetName()
	if err != nil {
		return nil, err
	}

	r.SetCacheControl()

	r.RemoveEmptyHeaders()

	assetID := p.AssetNameToAssetID(assetName)

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
		assetID := p.AssetNameToAssetID(assetItem.AssetName)
		u, err := p.storage.PresignGetObject(assetID)
		if err != nil {
			return nil, err
		}
		r.Assets[i].URL = u.String()
	}
	return r, nil
}

func (p *providerImpl) List(r *ListObjectsRequest) (*ListObjectsResponse, error) {
	if r.Prefix != "" {
		r.Prefix = p.AssetNameToAssetID(r.Prefix)
	}
	// 1000 is the greatest common page size.
	r.PageSize = 1000

	resp, err := p.storage.ListObjects(r)
	if err != nil {
		return nil, err
	}

	for i, assetItem := range resp.Assets {
		resp.Assets[i].AssetName = p.AssetIDToAssetName(assetItem.AssetName)
	}

	return resp, nil
}

func (p *providerImpl) Delete(name string) error {
	assetID := p.AssetNameToAssetID(name)
	err := p.storage.DeleteObject(assetID)
	if err != nil {
		return err
	}
	return nil
}

func (p *providerImpl) RewriteGetURL(u *url.URL, name string) (*url.URL, bool, error) {
	assetID := p.AssetNameToAssetID(name)
	return p.storage.RewriteGetURL(u, assetID)
}

func (p *providerImpl) ProprietaryToStandard(header http.Header) http.Header {
	return p.storage.ProprietaryToStandard(header)
}

func (p *providerImpl) AccessType(header http.Header) AccessType {
	return p.storage.AccessType(header)
}
