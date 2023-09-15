package webapp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
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
	GetToken(
		appID config.AppID,
		tokenID string,
		consume bool,
	) (*tester.TesterToken, error)
}

type TesterTokenIssuer interface {
	IssueTokensForAuthorizationCode(
		client *config.OAuthClientConfig,
		r protocol.TokenRequest,
	) (protocol.TokenResponse, error)
	IssueAppSessionToken(refreshToken string) (string, *oauth.AppSessionToken, error)
}

type TesterHandler struct {
	AppID                   config.AppID
	ControllerFactory       ControllerFactory
	OauthEndpointsProvider  oauth.EndpointsProvider
	TesterEndpointsProvider tester.EndpointsProvider
	TesterTokenStore        TesterTokenStore
	TesterTokenIssuer       TesterTokenIssuer
	OAuthClientResolver     *oauthclient.Resolver
}

type TesterState struct {
	Token string `json:"token"`
}

func (h *TesterHandler) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func (h *TesterHandler) triggerAuth(token string, w http.ResponseWriter, r *http.Request) error {
	testerToken, err := h.TesterTokenStore.GetToken(h.AppID, token, false)
	if errors.Is(err, tester.ErrTokenNotFound) {
		h.notFound(w, r)
		return nil
	}
	if err != nil {
		return err
	}
	state := &TesterState{
		Token: testerToken.TokenID,
	}
	stateBytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	stateb64 := base64.RawURLEncoding.EncodeToString(stateBytes)
	q := url.Values{}
	q.Set("redirect_uri", h.TesterEndpointsProvider.TesterURL().String())
	q.Set("scope", strings.Join([]string{
		"openid", "offline_access", "https://authgear.com/scopes/full-access",
	}, " "))
	q.Set("response_type", "code")
	q.Set("client_id", tester.ClientIDTester)
	q.Set("x_sso_enabled", "false")
	q.Set("state", stateb64)
	q.Set("code_challenge_method", testerToken.PKCEVerifier.CodeChallengeMethod)
	q.Set("code_challenge", testerToken.PKCEVerifier.Challenge())

	redirectTo := h.OauthEndpointsProvider.AuthorizeEndpointURL()
	redirectTo.RawQuery = q.Encode()
	http.Redirect(w, r, redirectTo.String(), http.StatusFound)

	return nil
}

func (h *TesterHandler) doCodeExchange(code string, stateb64 string, w http.ResponseWriter, r *http.Request) error {
	statejson, err := base64.RawURLEncoding.DecodeString(stateb64)
	var state *TesterState
	err = json.Unmarshal(statejson, &state)
	if err != nil {
		return err
	}

	testerToken, err := h.TesterTokenStore.GetToken(h.AppID, state.Token, true)
	if errors.Is(err, tester.ErrTokenNotFound) {
		h.notFound(w, r)
		return nil
	}
	client := h.OAuthClientResolver.ResolveClient(tester.ClientIDTester)
	tokenRequest := protocol.TokenRequest{}
	tokenRequest["code"] = code
	tokenRequest["code_verifier"] = testerToken.PKCEVerifier.CodeVerifier
	tokenRequest["redirect_uri"] = h.TesterEndpointsProvider.TesterURL().String()

	tokenResp, err := h.TesterTokenIssuer.IssueTokensForAuthorizationCode(
		client,
		tokenRequest,
	)

	if err != nil {
		return err
	}

	refreshToken, ok := tokenResp["refresh_token"].(string)
	if !ok {
		return fmt.Errorf("tester: refresh_token is not string")
	}
	fmt.Println("Refresh Token", refreshToken)

	// appSessionToken, _, err := h.TesterTokenIssuer.IssueAppSessionToken(refreshToken)

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

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code != "" && state != "" {
			return h.doCodeExchange(code, state, w, r)
		}

		h.notFound(w, r)
		return nil
	})

}

var _ http.Handler = &TesterHandler{}
