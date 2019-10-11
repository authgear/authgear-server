package cloudstorage

import (
	"fmt"
)

type MockProvider struct {
	PresignUploadResponse *PresignUploadResponse
}

var _ Provider = &MockProvider{}

func (p *MockProvider) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	return p.PresignUploadResponse, nil
}

func (p *MockProvider) Sign(r *SignRequest) (*SignRequest, error) {
	for i, assetItem := range r.Assets {
		r.Assets[i].URL = fmt.Sprintf("http://example/%s", assetItem.AssetID)
	}
	return r, nil
}
