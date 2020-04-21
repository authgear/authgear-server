package flows

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type TokenResult struct {
	Token string
}

type WebAppResult struct {
	Cookies []*http.Cookie
}

type AuthResult struct {
	Cookies  []*http.Cookie
	Response *model.AuthResponse
}

func (r *AuthResult) WriteResponse(rw http.ResponseWriter) {
	for _, c := range r.Cookies {
		corehttp.UpdateCookie(rw, c)
	}

	handler.WriteResponse(rw, handler.APIResponse{Result: r.Response})
}
