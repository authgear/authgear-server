package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Writer interface {
	WriteSession(rw http.ResponseWriter, resp *model.AuthResponse)
	ClearSession(rw http.ResponseWriter)
}
