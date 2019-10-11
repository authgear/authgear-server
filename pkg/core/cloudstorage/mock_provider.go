package cloudstorage

type MockProvider struct {
	PresignUploadResponse *PresignUploadResponse
}

var _ Provider = &MockProvider{}

func (p *MockProvider) PresignPutRequest(r *PresignUploadRequest) (*PresignUploadResponse, error) {
	return p.PresignUploadResponse, nil
}
