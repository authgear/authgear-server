package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type MockWriter struct {
}

func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

func (w *MockWriter) WriteSession(rw http.ResponseWriter, resp *model.AuthResponse) {
}

func (w *MockWriter) ClearSession(rw http.ResponseWriter) {
}
