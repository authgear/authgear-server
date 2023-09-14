package webapp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebTesterHTML = template.RegisterHTML(
	"web/tester.html",
	components...,
)

func ConfigureTesterRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET", "POST").
		WithPathPattern("/tester")
}

type TesterTokenStore interface {
	ConsumeToken(
		appID config.AppID,
		tokenID string,
	) (*tester.TesterToken, error)
}

type TesterHandler struct {
	AppID                   config.AppID
	ControllerFactory       ControllerFactory
	OauthEndpointsProvider  oauth.EndpointsProvider
	TesterEndpointsProvider tester.EndpointsProvider
	TesterTokenStore        TesterTokenStore
}

func (h *TesterHandler) triggerAuth(token string, w http.ResponseWriter, r *http.Request) error {
	testerToken, err := h.TesterTokenStore.ConsumeToken(h.AppID, token)
	if errors.Is(err, tester.ErrTokenNotFound) {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return nil
	}
	if err != nil {
		return err
	}
	stateMap := map[string]string{}
	stateMap["return_uri"] = testerToken.ReturnURI
	stateBytes, err := json.Marshal(stateMap)
	if err != nil {
		return err
	}
	stateb64 := make([]byte, base64.RawURLEncoding.EncodedLen(len(stateBytes)))
	base64.RawURLEncoding.Encode(stateb64, stateBytes)
	q := url.Values{}
	q.Set("redirect_uri", h.TesterEndpointsProvider.TesterURL().String())
	q.Set("scope", strings.Join([]string{
		"offline_access", "https://authgear.com/scopes/full-access",
	}, " "))
	q.Set("response_type", "code")
	q.Set("client_id", "tester")
	q.Set("x_sso_enabled", "false")
	q.Set("state", string(stateb64))

	redirectTo := h.OauthEndpointsProvider.AuthorizeEndpointURL()
	redirectTo.RawQuery = q.Encode()
	http.Redirect(w, r, redirectTo.String(), http.StatusFound)

	return nil
}

func (h *TesterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		token := r.URL.Query().Get("token")
		if token != "" {
			return h.triggerAuth(token, w, r)
		}
		return nil
	})

}

var _ http.Handler = &TesterHandler{}
