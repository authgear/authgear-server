package presign

import (
	"net/http"
)

type MockProvider struct{}

func (p *MockProvider) Presign(r *http.Request) {
}

func (p *MockProvider) Verify(r *http.Request) error {
	return nil
}
