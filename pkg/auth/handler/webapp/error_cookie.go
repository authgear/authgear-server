package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/web"
)

type ErrorCookie interface {
	GetError(r *http.Request) (*web.ErrorState, bool)
}
