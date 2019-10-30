package presign

import (
	"net/http"
	"time"
)

type MockProvider struct{}

func (p *MockProvider) Presign(r *http.Request, expires time.Duration) {
}

func (p *MockProvider) Verify(r *http.Request) error {
	return nil
}
