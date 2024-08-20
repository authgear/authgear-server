package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/pkce"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

//go:generate mockgen -source=handler_authz.go -destination=handler_authz_mock_test.go -package handler_test

const (
	CodeResponseTypeElement          = "code"
	NoneResponseTypeElement          = "none"
	TokenResponseTypeElement         = "token"
	SettingsActonResponseTypeElement = "urn:authgear:params:oauth:response-type:settings-action"
	// nolint:gosec
	PreAuthenticatedURLResponseTypeElement = "urn:authgear:params:oauth:response-type:pre-authenticated-url"
)

var (
	CodeResponseType                     = protocol.NewResponseType([]string{CodeResponseTypeElement})
	NoneResponseType                     = protocol.NewResponseType([]string{NoneResponseTypeElement})
	TokenResponseType                    = protocol.NewResponseType([]string{TokenResponseTypeElement})
	SettingsActonResponseType            = protocol.NewResponseType([]string{SettingsActonResponseTypeElement})
	PreAuthenticatedURLTokenResponseType = protocol.NewResponseType([]string{PreAuthenticatedURLResponseTypeElement, TokenResponseTypeElement})
)

// whitelistedResponseTypes is a list of response types that would be always allowed
// to all clients.
var whitelistedResponseTypes = []protocol.ResponseType{
	CodeResponseType,
	NoneResponseType,
	TokenResponseType,
	SettingsActonResponseType,
	PreAuthenticatedURLTokenResponseType,
}

const CodeGrantValidDuration = duration.Short
const SettingsActionGrantValidDuration = duration.Short

type UIInfoResolver interface {
	ResolveForAuthorizationEndpoint(client *config.OAuthClientConfig, req protocol.AuthorizationRequest) (*oidc.UIInfo, *oidc.UIInfoByProduct, error)
}

type AuthenticationInfoResolver interface {
	GetAuthenticationInfoID(req *http.Request) (string, bool)
}

type UIURLBuilder interface {
	BuildAuthenticationURL(client *config.OAuthClientConfig, r protocol.AuthorizationRequest, e *oauthsession.Entry) (*url.URL, error)
	BuildSettingsActionURL(client *config.OAuthClientConfig, r protocol.AuthorizationRequest, e *oauthsession.Entry) (*url.URL, error)
}

type AppSessionTokenService interface {
	Handle(input oauth.AppSessionTokenInput) (httputil.Result, error)
}

type AuthenticationInfoService interface {
	Get(entryID string) (*authenticationinfo.Entry, error)
	Delete(entryID string) error
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type OAuthSessionService interface {
	Save(entry *oauthsession.Entry) (err error)
	Get(entryID string) (*oauthsession.Entry, error)
	Delete(entryID string) error
}

type AuthorizationService interface {
	GetByID(id string) (*oauth.Authorization, error)
	CheckAndGrant(
		clientID string,
		userID string,
		scopes []string,
	) (*oauth.Authorization, error)
	Check(
		clientID string,
		userID string,
		scopes []string,
	) (*oauth.Authorization, error)
}

type AuthorizationHandlerLogger struct{ *log.Logger }

func NewAuthorizationHandlerLogger(lf *log.Factory) AuthorizationHandlerLogger {
	return AuthorizationHandlerLogger{lf.New("oauth-authz")}
}

type AuthorizationHandler struct {
	Context               context.Context
	AppID                 config.AppID
	Config                *config.OAuthConfig
	AccountDeletionConfig *config.AccountDeletionConfig
	HTTPConfig            *config.HTTPConfig
	HTTPProto             httputil.HTTPProto
	HTTPOrigin            httputil.HTTPOrigin
	AppDomains            config.AppDomains
	Logger                AuthorizationHandlerLogger

	UIURLBuilder                    UIURLBuilder
	UIInfoResolver                  UIInfoResolver
	AuthenticationInfoResolver      AuthenticationInfoResolver
	Authorizations                  AuthorizationService
	ValidateScopes                  ScopesValidator
	AppSessionTokenService          AppSessionTokenService
	AuthenticationInfoService       AuthenticationInfoService
	Clock                           clock.Clock
	Cookies                         CookieManager
	OAuthSessionService             OAuthSessionService
	CodeGrantService                CodeGrantService
	SettingsActionGrantService      SettingsActionGrantService
	ClientResolver                  OAuthClientResolver
	PreAuthenticatedURLTokenService PreAuthenticatedURLTokenService
	IDTokenIssuer                   IDTokenIssuer
}

func (h *AuthorizationHandler) Handle(r protocol.AuthorizationRequest) httputil.Result {
	client := resolveClient(h.ClientResolver, r)
	if client == nil {
		return authorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("unauthorized_client", "invalid client ID"),
		}
	}

	originWhitelist := []string{}
	if r.ResponseType().Equal(PreAuthenticatedURLTokenResponseType) {
		originWhitelist = client.PreAuthenticatedURLAllowedOrigins
	}

	redirectURI, errResp := parseRedirectURI(client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, originWhitelist, r)
	if errResp != nil {
		return authorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     errResp,
		}
	}

	result, err := h.doHandle(redirectURI, client, r)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		resultErr := authorizationResultError{
			RedirectURI:  redirectURI,
			ResponseMode: r.ResponseMode(),
		}
		if errors.As(err, &oauthError) {
			resultErr.Response = oauthError.Response
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
			resultErr.InternalError = true
		}
		state := r.State()
		if state != "" {
			resultErr.Response.State(r.State())
		}
		result = resultErr
	}

	return result
}

func (h *AuthorizationHandler) HandleConsentWithoutUserConsent(req *http.Request) (httputil.Result, *ConsentRequired) {
	result, consentRequired := h.doHandleConsent(req, false)
	return result, consentRequired
}

func (h *AuthorizationHandler) HandleConsentWithUserConsent(req *http.Request) httputil.Result {
	result, _ := h.doHandleConsent(req, true)
	return result
}

func (h *AuthorizationHandler) HandleConsentWithUserCancel(req *http.Request) httputil.Result {
	consentRequest, err := h.prepareConsentRequest(req)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		var resultErr httputil.Result

		if errors.As(err, &oauthError) {
			resultErr = h.prepareConsentErrInvalidOAuthResponse(req, *oauthError)
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr = authorizationResultError{
				// Don't redirect for those unexpected errors
				// e.g. oauth session expire or invalid client_id, redirect_uri
				RedirectURI:   nil,
				Response:      protocol.NewErrorResponse("server_error", "internal server error"),
				InternalError: true,
			}
		}

		return resultErr
	}

	oauthSessionEntry := consentRequest.OAuthSessionEntry
	authInfoEntry := consentRequest.AuthInfoEntry
	authzReq := oauthSessionEntry.T.AuthorizationRequest
	redirectURI := consentRequest.RedirectURI
	// delete oauth session and auth info with best effort
	// don't block the user in case of failure
	err = h.OAuthSessionService.Delete(oauthSessionEntry.ID)
	if err != nil {
		h.Logger.WithError(err).Error("failed to consume oauth session")
	}
	err = h.AuthenticationInfoService.Delete(authInfoEntry.ID)
	if err != nil {
		h.Logger.WithError(err).Error("failed to consume authentication info")
	}

	resultErr := authorizationResultError{
		ResponseMode: authzReq.ResponseMode(),
		RedirectURI:  redirectURI,
		Response:     protocol.NewErrorResponse("access_denied", "authorization denied"),
	}
	state := authzReq.State()
	if state != "" {
		resultErr.Response.State(authzReq.State())
	}
	return resultErr
}

type ConsentRequired struct {
	UserID string
	Scopes []string
	Client *config.OAuthClientConfig
}

func (h *AuthorizationHandler) doHandleConsent(req *http.Request, withUserConsent bool) (httputil.Result, *ConsentRequired) {
	consentRequest, err := h.prepareConsentRequest(req)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		var resultErr httputil.Result

		if errors.As(err, &oauthError) {
			resultErr = h.prepareConsentErrInvalidOAuthResponse(req, *oauthError)
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr = authorizationResultError{
				// Don't redirect for those unexpected errors
				// e.g. oauth session expire or invalid client_id, redirect_uri
				RedirectURI:   nil,
				Response:      protocol.NewErrorResponse("server_error", "internal server error"),
				InternalError: true,
			}
		}

		return resultErr, nil
	}

	oauthSessionEntry := consentRequest.OAuthSessionEntry
	redirectURI := consentRequest.RedirectURI
	client := consentRequest.Client
	authInfoEntry := consentRequest.AuthInfoEntry

	authzReq := oauthSessionEntry.T.AuthorizationRequest
	autoGrantAuthz := client.IsFirstParty()
	grantAuthz := autoGrantAuthz || withUserConsent

	result, err := h.doHandleConsentRequest(redirectURI, client, authzReq, authInfoEntry.T, req, grantAuthz)
	if err != nil {
		if !grantAuthz && IsConsentRequiredError(err) {
			return nil, &ConsentRequired{
				UserID: authInfoEntry.T.UserID,
				Scopes: authzReq.Scope(),
				Client: client,
			}
		}

		var oauthError *protocol.OAuthProtocolError
		resultErr := authorizationResultError{
			RedirectURI:  redirectURI,
			ResponseMode: authzReq.ResponseMode(),
		}
		if errors.As(err, &oauthError) {
			resultErr.Response = oauthError.Response
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
			resultErr.InternalError = true
		}
		state := authzReq.State()
		if state != "" {
			resultErr.Response.State(authzReq.State())
		}
		result = resultErr
	} else {
		// delete oauth session with best effort
		// don't block the user in case of failure
		err = h.OAuthSessionService.Delete(oauthSessionEntry.ID)
		if err != nil {
			h.Logger.WithError(err).Error("failed to consume oauth session")
		}
		err = h.AuthenticationInfoService.Delete(authInfoEntry.ID)
		if err != nil {
			h.Logger.WithError(err).Error("failed to consume authentication info")
		}
	}

	return result, nil
}

func (h *AuthorizationHandler) getAuthenticationInfoEntry(req *http.Request) (*authenticationinfo.Entry, error) {
	id, ok := h.AuthenticationInfoResolver.GetAuthenticationInfoID(req)
	if !ok {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	entry, err := h.AuthenticationInfoService.Get(id)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

type consentRequest struct {
	OAuthSessionEntry *oauthsession.Entry
	AuthInfoEntry     *authenticationinfo.Entry
	RedirectURI       *url.URL
	Client            *config.OAuthClientConfig
}

func (h *AuthorizationHandler) prepareConsentRequest(req *http.Request) (*consentRequest, error) {
	authInfoEntry, err := h.getAuthenticationInfoEntry(req)
	if err != nil {
		if errors.Is(err, authenticationinfo.ErrNotFound) {
			err = protocol.NewError("invalid_request", "authentication expired")
		}
		return nil, err
	}

	entry, err := h.OAuthSessionService.Get(authInfoEntry.OAuthSessionID)
	if err != nil {
		if errors.Is(err, oauthsession.ErrNotFound) {
			err = protocol.NewError("invalid_request", "oauth session expired")
		}
		return nil, err
	}

	r := entry.T.AuthorizationRequest

	client := resolveClient(h.ClientResolver, r)
	if client == nil {
		err = protocol.NewError("unauthorized_client", "invalid client ID")
		return nil, err
	}

	uiInfo, _, err := h.UIInfoResolver.ResolveForAuthorizationEndpoint(client, r)
	if err != nil {
		return nil, err
	}

	uiParam := uiInfo.ToUIParam()
	// Restore uiparam into context.
	uiparam.WithUIParam(h.Context, &uiParam)

	redirectURI, errResp := parseRedirectURI(client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, []string{}, r)
	if errResp != nil {
		err = protocol.NewErrorWithErrorResponse(errResp)
		return nil, err
	}

	return &consentRequest{
		OAuthSessionEntry: entry,
		AuthInfoEntry:     authInfoEntry,
		RedirectURI:       redirectURI,
		Client:            client,
	}, nil
}

// nolint: gocognit
func (h *AuthorizationHandler) doHandle(
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) (httputil.Result, error) {
	if err := h.validateRequest(client, r); err != nil {
		return nil, err
	}

	if r.ResponseType().Equal(PreAuthenticatedURLTokenResponseType) {
		return h.doHandlePreAuthenticatedURL(redirectURI, client, r)
	}

	err := h.ValidateScopes(client, r.Scope())
	if err != nil {
		return nil, err
	}

	// create oauth session and redirect to the web app
	oauthSessionEntry := oauthsession.NewEntry(oauthsession.T{
		AuthorizationRequest: r,
	})
	err = h.OAuthSessionService.Save(oauthSessionEntry)
	if err != nil {
		return nil, err
	}

	if r.ResponseType().Equal(SettingsActonResponseType) {
		redirectURI, err = h.UIURLBuilder.BuildSettingsActionURL(client, r, oauthSessionEntry)
		if err != nil {
			return nil, err
		}
	}

	loginHintString, loginHintOk := r.LoginHint()
	// Handle app session token here, and return here.
	// Anonymous user promotion is handled by the normal flow below.
	if loginHintOk {
		loginHint, err := oauth.ParseLoginHint(loginHintString)
		if err != nil {
			return nil, protocol.NewError("invalid_request", err.Error())
		}

		if loginHint.Type == oauth.LoginHintTypeAppSessionToken {
			result, err := h.AppSessionTokenService.Handle(oauth.AppSessionTokenInput{
				AppSessionToken: loginHint.AppSessionToken,
				RedirectURI:     redirectURI.String(),
			})
			if err != nil {
				return nil, protocol.NewError("invalid_request", err.Error())
			}
			return result, nil
		}
	}

	uiInfo, uiInfoByProduct, err := h.UIInfoResolver.ResolveForAuthorizationEndpoint(client, r)
	if err != nil {
		return nil, err
	}
	idToken := uiInfoByProduct.IDToken
	idTokenHintSID := uiInfoByProduct.IDTokenHintSID

	// Handle prompt!=none
	// We must return here.
	if !slice.ContainsString(uiInfo.Prompt, "none") {
		endpoint, err := h.UIURLBuilder.BuildAuthenticationURL(client, r, oauthSessionEntry)
		if apierrors.IsKind(err, oidc.ErrInvalidCustomURI) {
			return nil, protocol.NewError("invalid_request", err.Error())
		} else if err != nil {
			return nil, err
		}

		resp := &httputil.ResultRedirect{
			Cookies: []*http.Cookie{
				h.Cookies.ValueCookie(oauthsession.UICookieDef, oauthSessionEntry.ID),
			},
			URL: endpoint.String(),
		}
		return resp, nil
	}

	// Handle prompt=none
	var resolvedSession session.ResolvedSession
	if s := session.GetSession(h.Context); s != nil {
		resolvedSession = s
	}
	// Ignore any session that is not allow to be used here
	if !oauth.ContainsAllScopes(oauth.SessionScopes(resolvedSession), []string{oauth.PreAuthenticatedURLScope}) {
		resolvedSession = nil
	}
	if resolvedSession == nil || (idToken != nil && resolvedSession.GetAuthenticationInfo().UserID != idToken.Subject()) {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	authenticationInfo := resolvedSession.CreateNewAuthenticationInfoByThisSession()
	autoGrantAuthz := client.IsFirstParty()

	sessionType := resolvedSession.SessionType()
	sessionID := resolvedSession.SessionID()

	result, err := h.finish(redirectURI, r, sessionType, sessionID, authenticationInfo, idTokenHintSID, nil, autoGrantAuthz)
	if err != nil {
		if errors.Is(err, oauth.ErrAuthorizationNotFound) {
			return nil, protocol.NewError("access_denied", "authorization required")
		}
		if errors.Is(err, oauth.ErrAuthorizationScopesNotGranted) {
			return nil, protocol.NewError("access_denied", "requested scopes are not granted")
		}
		return nil, err
	}
	return result, nil
}

func (h *AuthorizationHandler) doHandlePreAuthenticatedURL(
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) (httputil.Result, error) {
	idTokenHint, ok := r.IDTokenHint()
	if !ok {
		panic("cannot get id_token_hint, the request should be validated")
	}
	idToken, err := h.IDTokenIssuer.VerifyIDToken(idTokenHint)
	if err != nil {
		return nil, protocol.NewError("invalid_request", "invalid id_token_hint")
	}
	var sidInt interface{}
	if sidInt, ok = idToken.Get(string(model.ClaimSID)); !ok {
		return nil, protocol.NewError("invalid_request", "required sid in id_token_hint")
	}
	var sid string
	if sid, ok = sidInt.(string); !ok {
		return nil, protocol.NewError("invalid_request", "sid is not a string in id_token_hint")
	}
	_, sessionID, ok := oidc.DecodeSID(sid)
	if !ok {
		return nil, protocol.NewError("invalid_request", "invalid sid format id_token_hint")
	}

	accessToken, err := h.PreAuthenticatedURLTokenService.ExchangeForAccessToken(
		client,
		sessionID,
		r.PreAuthenticatedURLToken(),
	)
	if err != nil {
		if errors.Is(err, oauth.ErrUnmatchedClient) {
			return nil, protocol.NewError("invalid_request", "incorrect client_id")
		}
		if errors.Is(err, oauth.ErrUnmatchedSession) {
			return nil, protocol.NewError("invalid_request", "incorrect sid in id_token_hint")
		}
		if errors.Is(err, oauth.ErrGrantNotFound) {
			return nil, protocol.NewError("invalid_request", "invalid x_pre_authenticated_url_token")
		}
		return nil, err
	}
	cookie := h.Cookies.ValueCookie(session.AppAccessTokenCookieDef, accessToken)

	resp := protocol.AuthorizationResponse{}
	state := r.State()
	if state != "" {
		resp.State(r.State())
	}
	return authorizationResultCode{
		RedirectURI:  redirectURI,
		ResponseMode: r.ResponseMode(),
		Response:     resp,
		Cookies:      []*http.Cookie{cookie},
	}, nil

}

func (h *AuthorizationHandler) finish(
	redirectURI *url.URL,
	r protocol.AuthorizationRequest,
	sessionType session.Type,
	sessionID string,
	authenticationInfo authenticationinfo.T,
	idTokenHintSID string,
	cookies []*http.Cookie,
	grantAuthz bool,
) (httputil.Result, error) {
	var authz *oauth.Authorization
	var err error
	if grantAuthz {
		authz, err = h.Authorizations.CheckAndGrant(
			r.ClientID(),
			authenticationInfo.UserID,
			r.Scope(),
		)
	} else {
		authz, err = h.Authorizations.Check(
			r.ClientID(),
			authenticationInfo.UserID,
			r.Scope(),
		)
	}
	if err != nil {
		return nil, err
	}

	resp := protocol.AuthorizationResponse{}
	responseType := r.ResponseType()
	switch {
	case responseType.Equal(SettingsActonResponseType):
		idpSessionID := ""
		if sessionType == session.TypeIdentityProvider {
			idpSessionID = sessionID
		}
		err = h.generateSettingsActionResponse(redirectURI.String(), idpSessionID, authenticationInfo, idTokenHintSID, r, authz, resp)
		if err != nil {
			return nil, err
		}

	case responseType.Equal(CodeResponseType):
		err = h.generateCodeResponse(redirectURI.String(), sessionType, sessionID, authenticationInfo, idTokenHintSID, r, authz, resp)
		if err != nil {
			return nil, err
		}

	case responseType.Equal(NoneResponseType):
		break

	default:
		panic("oauth: unexpected response type")
	}

	state := r.State()
	if state != "" {
		resp.State(r.State())
	}

	return authorizationResultCode{
		RedirectURI:  redirectURI,
		ResponseMode: r.ResponseMode(),
		Response:     resp,
		Cookies:      cookies,
	}, nil
}

func (h *AuthorizationHandler) doHandleConsentRequest(
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
	authenticationInfo authenticationinfo.T,
	req *http.Request,
	grantAuthz bool,
) (httputil.Result, error) {
	if err := h.validateRequest(client, r); err != nil {
		return nil, err
	}

	err := h.ValidateScopes(client, r.Scope())
	if err != nil {
		return nil, err
	}

	_, uiInfoByProduct, err := h.UIInfoResolver.ResolveForAuthorizationEndpoint(client, r)
	if err != nil {
		return nil, err
	}
	idTokenHintSID := uiInfoByProduct.IDTokenHintSID

	sessionID := ""
	var sessionType session.Type = ""

	if authenticationInfo.AuthenticatedBySessionID != "" {
		sessionID = authenticationInfo.AuthenticatedBySessionID
		sessionType = session.Type(authenticationInfo.AuthenticatedBySessionType)
	}

	return h.finish(redirectURI, r, sessionType, sessionID, authenticationInfo, idTokenHintSID, []*http.Cookie{}, grantAuthz)
}

func (h *AuthorizationHandler) validatePreAuthenticatedURLTokenRequest(
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) error {
	if len(r.Prompt()) != 1 || r.Prompt()[0] != "none" {
		return protocol.NewError("invalid_request", "only 'prompt=none' is supported when using pre-authenticated url")
	}
	if idTokenHint, ok := r.IDTokenHint(); !ok || idTokenHint == "" {
		return protocol.NewError("invalid_request", "id_token_hint is required when using pre-authenticated url")
	}
	if r.PreAuthenticatedURLToken() == "" {
		return protocol.NewError("invalid_request", "x_pre_authenticated_url_token is required when using pre-authenticated url")
	}
	if r.ResponseMode() != "cookie" {
		return protocol.NewError("invalid_request", "only 'response_mode=cookie' is supported when using pre-authenticated url")
	}
	return nil
}

// nolint:gocognit
func (h *AuthorizationHandler) validateRequest(
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) error {
	ok := false
	responseType := r.ResponseType()
	for _, respType := range whitelistedResponseTypes {
		if respType.Equal(responseType) {
			ok = true
			break
		}
	}
	if !ok {
		return protocol.NewError("unauthorized_client", "response type is not allowed for this client")
	}

	if slice.ContainsString(r.Prompt(), "none") {
		if len(r.Prompt()) != 1 {
			return protocol.NewError("invalid_request", "prompt cannot have other values when none is set")
		}
		if r.HasMaxAge() {
			return protocol.NewError("invalid_request", "max_age could imply prompt=login so max_age cannot be present when prompt=none")
		}
	}

	requireScope := func() error {
		if len(r.Scope()) == 0 {
			return protocol.NewError("invalid_request", "scope is required")
		}
		return nil
	}

	switch {
	case responseType.Equal(SettingsActonResponseType):
		if r.SettingsAction() == "delete_account" {
			if !h.AccountDeletionConfig.ScheduledByEndUserEnabled {
				return protocol.NewError("invalid_request", "account deletion by end user is disabled")
			}
		}
		fallthrough
	case responseType.Equal(CodeResponseType):
		if err := requireScope(); err != nil {
			return err
		}
		if client.IsPublic() {
			if r.CodeChallenge() == "" {
				return protocol.NewError("invalid_request", "PKCE code challenge is required for public clients")
			}
		}
		if r.CodeChallenge() != "" && r.CodeChallengeMethod() != pkce.CodeChallengeMethodS256 {
			return protocol.NewError("invalid_request", "only 'S256' PKCE transform is supported")
		}
	case responseType.Equal(NoneResponseType):
		if err := requireScope(); err != nil {
			return err
		}
	case responseType.Equal(PreAuthenticatedURLTokenResponseType):
		if err := h.validatePreAuthenticatedURLTokenRequest(client, r); err != nil {
			return err
		}
	default:
		return protocol.NewError("unsupported_response_type", fmt.Sprintf("response_type: %v is not supported", responseType.Raw))
	}

	if r.SSOEnabled() && client != nil && client.MaxConcurrentSession == 1 {
		return protocol.NewError("invalid_request", "'sso_enabled' must be false if config 'x_max_concurrent_session' is 1")
	}

	return nil
}

func (h *AuthorizationHandler) generateCodeResponse(
	redirectURI string,
	sessionType session.Type,
	sessionID string,
	authenticationInfo authenticationinfo.T,
	idTokenHintSID string,
	r protocol.AuthorizationRequest,
	authz *oauth.Authorization,
	resp protocol.AuthorizationResponse,
) error {
	code, _, err := h.CodeGrantService.CreateCodeGrant(&CreateCodeGrantOptions{
		Authorization:        authz,
		SessionType:          sessionType,
		SessionID:            sessionID,
		AuthenticationInfo:   authenticationInfo,
		IDTokenHintSID:       idTokenHintSID,
		RedirectURI:          redirectURI,
		AuthorizationRequest: r,
		DPoPJKT:              r.DPoPJKT(),
	})
	if err != nil {
		return err
	}

	resp.Code(code)
	return nil
}

func (h *AuthorizationHandler) generateSettingsActionResponse(
	redirectURI string,
	idpSessionID string,
	authenticationInfo authenticationinfo.T,
	idTokenHintSID string,
	r protocol.AuthorizationRequest,
	authz *oauth.Authorization,
	resp protocol.AuthorizationResponse,
) error {
	code, _, err := h.SettingsActionGrantService.CreateSettingsActionGrant(&CreateSettingsActionGrantOptions{
		RedirectURI:          redirectURI,
		AuthorizationRequest: r,
	})
	if err != nil {
		return err
	}

	resp.Code(code)
	return nil
}

func (h *AuthorizationHandler) prepareConsentErrInvalidOAuthResponse(req *http.Request, oauthError protocol.OAuthProtocolError) httputil.Result {
	resultErr := authorizationResultError{
		Response: oauthError.Response,
	}

	state := req.URL.Query().Get("state")
	if state != "" {
		resultErr.Response.State(state)
	}

	client := h.ClientResolver.ResolveClient(req.URL.Query().Get("client_id"))

	// Only redirect if oauth session is expired / not found
	// It mostly happens when user refresh the page or go back to the page after authenication
	if oauthError.Type() == "invalid_request" && client != nil {
		redirectURI, err := url.Parse(req.URL.Query().Get("redirect_uri"))
		if err == nil {
			err = validateRedirectURI(client, h.HTTPProto, h.HTTPOrigin, h.AppDomains, []string{}, redirectURI)
			if err == nil {
				resultErr.RedirectURI = redirectURI
			}
		}
	}

	return resultErr
}
