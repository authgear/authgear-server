package cloudstorage

import (
	"fmt"
	"net/http"
	"net/url"
)

type MockProvider struct {
	PresignUploadResponse *PresignUploadResponse
	ListObjectsResponse   *ListObjectsResponse
	GetURL                *url.URL
	GetAccessType         AccessType
}

var _ Provider = &MockProvider{}

func (p *MockProvider) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	return p.PresignUploadResponse, nil
}

func (p *MockProvider) Sign(scheme string, host string, r *SignRequest) (*SignRequest, error) {
	for i, assetItem := range r.Assets {
		r.Assets[i].URL = fmt.Sprintf("%s://%s/_asset/get/%s", scheme, host, assetItem.AssetName)
	}
	return r, nil
}

func (p *MockProvider) Verify(r *http.Request) error {
	return nil
}

func (p *MockProvider) PresignGetRequest(assetName string) (*url.URL, error) {
	return p.GetURL, nil
}

func (p *MockProvider) List(r *ListObjectsRequest) (*ListObjectsResponse, error) {
	return p.ListObjectsResponse, nil
}

func (p *MockProvider) Delete(name string) error {
	return nil
}

func (p *MockProvider) ProprietaryToStandard(header http.Header) http.Header {
	return header
}

func (p *MockProvider) AccessType(header http.Header) AccessType {
	return p.GetAccessType
}
