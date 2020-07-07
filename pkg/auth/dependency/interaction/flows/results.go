package flows

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/core/handler"
	"github.com/authgear/authgear-server/pkg/httputil"
)

type WebAppResult struct {
	Interaction *interaction.Interaction
	Cookies     []*http.Cookie
}

type AuthResult struct {
	Cookies  []*http.Cookie
	Response *model.AuthResponse
}

func (r *AuthResult) WriteResponse(rw http.ResponseWriter) {
	for _, c := range r.Cookies {
		httputil.UpdateCookie(rw, c)
	}

	handler.WriteResponse(rw, handler.APIResponse{Result: r.Response})
}
