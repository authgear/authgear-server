package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureSSOCallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/sso/oauth2/callback/:alias")
}

type SSOCallbackHandler struct {
	Database *db.Handle
	WebApp   WebAppService
}

type SSOCallbackInput struct {
	ProviderAlias string
	NonceSource   *http.Cookie

	State            string
	Code             string
	Scope            string
	Error            string
	ErrorDescription string
}

func (i *SSOCallbackInput) GetProviderAlias() string {
	return i.ProviderAlias
}

func (i *SSOCallbackInput) GetNonceSource() *http.Cookie {
	return i.NonceSource
}

func (i *SSOCallbackInput) GetCode() string {
	return i.Code
}

func (i *SSOCallbackInput) GetState() string {
	return i.State
}

func (i *SSOCallbackInput) GetScope() string {
	return i.Scope
}

func (i *SSOCallbackInput) GetError() string {
	return i.Error
}

func (i *SSOCallbackInput) GetErrorDescription() string {
	return i.ErrorDescription
}

var _ nodes.InputSelectIdentityOAuthUserInfo = &SSOCallbackInput{}

func (h *SSOCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nonceSource, _ := r.Cookie(webapp.CSRFCookieName)

	data := SSOCallbackInput{
		ProviderAlias: httproute.GetParam(r, "alias"),
		NonceSource:   nonceSource,

		State:            r.Form.Get("state"),
		Code:             r.Form.Get("code"),
		Scope:            r.Form.Get("scope"),
		Error:            r.Form.Get("error"),
		ErrorDescription: r.Form.Get("error_description"),
	}

	h.Database.WithTx(func() error {
		result, err := h.WebApp.PostInput(data.State, func() (input interface{}, err error) {
			input = &data
			return
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		result.WriteResponse(w, r)
		return nil
	})
}
