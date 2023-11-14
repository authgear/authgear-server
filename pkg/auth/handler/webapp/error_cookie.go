package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
)

type ErrorCookie interface {
	GetError(r *http.Request) (*webapp.ErrorState, bool)
}
