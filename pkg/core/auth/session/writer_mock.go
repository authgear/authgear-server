package session

import (
	"net/http"
)

type MockWriter struct {
}

func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

var _ Writer = &MockWriter{}

func (w *MockWriter) WriteSession(rw http.ResponseWriter, accessToken *string, mfaBearerToken *string) {
}

func (w *MockWriter) ClearSession(rw http.ResponseWriter) {
}
