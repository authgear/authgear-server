package session

import (
	"net/http"
)

type MockWriter struct {
}

func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

func (w *MockWriter) WriteSession(rw http.ResponseWriter, accessToken *string) {
}

func (w *MockWriter) ClearSession(rw http.ResponseWriter) {
}
