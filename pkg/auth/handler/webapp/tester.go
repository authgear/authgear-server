package webapp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
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

type TesterViewModel struct {
	ReturnURI    string
	UserInfoJson string
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

type TesterAppSessionTokenService interface {
	Exchange(appSessionToken string) (string, error)
}

type TesterCookieManager interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
}

type TesterUserInfoProvider interface {
	GetUserInfo(userID string, clientLike *oauth.ClientLike) (map[string]interface{}, error)
}

type TesterOfflineGrantStore interface {
	GetOfflineGrant(id string) (*oauth.OfflineGrant, error)
}

type TesterHandler struct {
	AppID                   config.AppID
	ControllerFactory       ControllerFactory
	OauthEndpointsProvider  oauth.EndpointsProvider
	TesterEndpointsProvider tester.EndpointsProvider
	TesterTokenStore        TesterTokenStore
	TesterTokenIssuer       TesterTokenIssuer
	OAuthClientResolver     *oauthclient.Resolver
	AppSessionTokenService  TesterAppSessionTokenService
	CookieManager           TesterCookieManager
	Renderer                Renderer
	BaseViewModel           *viewmodels.BaseViewModeler
	UserInfoProvider        TesterUserInfoProvider
	OfflineGrants           TesterOfflineGrantStore
}

var TesterScopes = []string{
	"openid", "offline_access", "https://authgear.com/scopes/full-access",
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
	q.Set("scope", strings.Join(TesterScopes, " "))
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

func (h *TesterHandler) getData(
	rw http.ResponseWriter,
	r *http.Request,
	testerToken *tester.TesterToken,
	userInfo map[string]interface{},
) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	userInfoJsonBytes, err := json.MarshalIndent(userInfo, "", "  ")
	if err != nil {
		return nil, err
	}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	testerViewModel := TesterViewModel{
		ReturnURI:    testerToken.ReturnURI,
		UserInfoJson: string(userInfoJsonBytes),
	}

	viewmodels.Embed(data, baseViewModel)
	viewmodels.Embed(data, testerViewModel)
	return data, nil
}

func (h *TesterHandler) doCodeExchange(code string, stateb64 string, w http.ResponseWriter, r *http.Request) error {
	statejson, err := base64.RawURLEncoding.DecodeString(stateb64)
	if err != nil {
		return err
	}
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

	appSessionToken, s, err := h.TesterTokenIssuer.IssueAppSessionToken(refreshToken)
	if err != nil {
		return err
	}

	offlineGrant, err := h.OfflineGrants.GetOfflineGrant(s.OfflineGrantID)
	if err != nil {
		return err
	}
	userID := offlineGrant.GetAuthenticationInfo().UserID

	appSession, err := h.AppSessionTokenService.Exchange(appSessionToken)
	if err != nil {
		return err
	}

	cookie := h.CookieManager.ValueCookie(session.AppSessionTokenCookieDef, appSession)
	httputil.UpdateCookie(w, cookie)

	userInfo, err := h.UserInfoProvider.GetUserInfo(userID, oauth.ClientClientLike(client, TesterScopes))
	if err != nil {
		return err
	}

	data, err := h.getData(w, r, testerToken, userInfo)
	if err != nil {
		return err
	}

	h.Renderer.RenderHTML(w, r, TemplateWebTesterHTML, data)

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
