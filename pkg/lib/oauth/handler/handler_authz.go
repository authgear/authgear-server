package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

//go:generate mockgen -source=handler_authz.go -destination=handler_authz_mock_test.go -package handler_test

const CodeGrantValidDuration = duration.Short

type UIInfoResolver interface {
	ResolveForAuthorizationEndpoint(client *config.OAuthClientConfig, req protocol.AuthorizationRequest) (*oidc.UIInfo, *oidc.UIInfoByProduct, error)
}

type UIURLBuilder interface {
	Build(client *config.OAuthClientConfig, r protocol.AuthorizationRequest) (*url.URL, error)
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
	Context    context.Context
	AppID      config.AppID
	Config     *config.OAuthConfig
	HTTPConfig *config.HTTPConfig
	Logger     AuthorizationHandlerLogger

	UIURLBuilder              UIURLBuilder
	UIInfoResolver            UIInfoResolver
	Authorizations            AuthorizationService
	ValidateScopes            ScopesValidator
	AppSessionTokenService    AppSessionTokenService
	AuthenticationInfoService AuthenticationInfoService
	Clock                     clock.Clock
	Cookies                   CookieManager
	OAuthSessionService       OAuthSessionService
	CodeGrantService          CodeGrantService
}

func (h *AuthorizationHandler) Handle(r protocol.AuthorizationRequest) httputil.Result {
	client := resolveClient(h.Config, r)
	if client == nil {
		return authorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("unauthorized_client", "invalid client ID"),
		}
	}
	redirectURI, errResp := parseRedirectURI(client, h.HTTPConfig, r)
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
		resultErr := authorizationResultError{
			// Don't redirect for those unexpected errors
			// e.g. oauth session expire or invalid client_id, redirect_uri
			RedirectURI: nil,
		}
		if errors.As(err, &oauthError) {
			resultErr.Response = oauthError.Response
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
			resultErr.InternalError = true
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
		Cookies: []*http.Cookie{
			h.Cookies.ClearCookie(authenticationinfo.CookieDef),
			h.Cookies.ClearCookie(oauthsession.CookieDef),
		},
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
		resultErr := authorizationResultError{
			// Don't redirect for those unexpected errors
			// e.g. oauth session expire or invalid client_id, redirect_uri
			RedirectURI: nil,
		}
		if errors.As(err, &oauthError) {
			resultErr.Response = oauthError.Response
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
			resultErr.InternalError = true
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
	cookie, err := h.Cookies.GetCookie(req, authenticationinfo.CookieDef)
	if err != nil {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	entry, err := h.AuthenticationInfoService.Get(cookie.Value)
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
	cookie, err := h.Cookies.GetCookie(req, oauthsession.CookieDef)
	if err != nil {
		err = protocol.NewError("invalid_request", "missing oauth session")
		return nil, err
	}

	entry, err := h.OAuthSessionService.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, oauthsession.ErrNotFound) {
			err = protocol.NewError("invalid_request", "oauth session expired")
		}
		return nil, err
	}

	r := entry.T.AuthorizationRequest

	client := resolveClient(h.Config, r)
	if client == nil {
		err = protocol.NewError("unauthorized_client", "invalid client ID")
		return nil, err
	}

	redirectURI, errResp := parseRedirectURI(client, h.HTTPConfig, r)
	if errResp != nil {
		err = protocol.NewErrorWithErrorResponse(errResp)
		return nil, err
	}

	authInfoEntry, err := h.getAuthenticationInfoEntry(req)
	if err != nil {
		if errors.Is(err, authenticationinfo.ErrNotFound) {
			err = protocol.NewError("invalid_request", "authentication expired")
		}
		return nil, err
	}

	return &consentRequest{
		OAuthSessionEntry: entry,
		AuthInfoEntry:     authInfoEntry,
		RedirectURI:       redirectURI,
		Client:            client,
	}, nil
}

// nolint: gocyclo
func (h *AuthorizationHandler) doHandle(
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) (httputil.Result, error) {
	if err := h.validateRequest(client, r); err != nil {
		return nil, err
	}

	err := h.ValidateScopes(client, r.Scope())
	if err != nil {
		return nil, err
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

	// create oauth session and redirect to the web app
	oauthSessionEntry := oauthsession.NewEntry(oauthsession.T{
		AuthorizationRequest: r,
	})
	err = h.OAuthSessionService.Save(oauthSessionEntry)
	if err != nil {
		return nil, err
	}

	// Handle prompt!=none
	// We must return here.
	if !slice.ContainsString(uiInfo.Prompt, "none") {
		endpoint, err := h.UIURLBuilder.Build(client, r)
		if apierrors.IsKind(err, oidc.ErrInvalidCustomURI) {
			return nil, protocol.NewError("invalid_request", err.Error())
		} else if err != nil {
			return nil, err
		}

		resp := &httputil.ResultRedirect{
			Cookies: []*http.Cookie{
				h.Cookies.ValueCookie(oauthsession.CookieDef, oauthSessionEntry.ID),
				h.Cookies.ValueCookie(oauthsession.UICookieDef, oauthSessionEntry.ID),
			},
			URL: endpoint.String(),
		}
		return resp, nil
	}

	// Handle prompt=none
	var idpSession session.Session
	if s := session.GetSession(h.Context); s != nil && s.SessionType() == session.TypeIdentityProvider {
		idpSession = s
	}
	if idpSession == nil || (idToken != nil && idpSession.GetAuthenticationInfo().UserID != idToken.Subject()) {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	authenticationInfo := idpSession.GetAuthenticationInfo()
	autoGrantAuthz := client.IsFirstParty()

	result, err := h.finish(redirectURI, r, idpSession.SessionID(), authenticationInfo, idTokenHintSID, nil, autoGrantAuthz)
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

func (h *AuthorizationHandler) finish(
	redirectURI *url.URL,
	r protocol.AuthorizationRequest,
	idpSessionID string,
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
	switch r.ResponseType() {
	case "code":
		err = h.generateCodeResponse(redirectURI.String(), idpSessionID, authenticationInfo, idTokenHintSID, r, authz, resp)
		if err != nil {
			return nil, err
		}

	case "none":
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
		Cookies: append(
			[]*http.Cookie{h.Cookies.ClearCookie(authenticationinfo.CookieDef)},
			cookies...,
		),
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

	var idpSessionID string
	if s := session.GetSession(h.Context); s != nil && s.SessionType() == session.TypeIdentityProvider {
		idpSessionID = s.SessionID()
	}

	return h.finish(redirectURI, r, idpSessionID, authenticationInfo, idTokenHintSID, []*http.Cookie{h.Cookies.ClearCookie(oauthsession.CookieDef)}, grantAuthz)
}

func (h *AuthorizationHandler) validateRequest(
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) error {
	allowedResponseTypes := client.ResponseTypes
	if len(allowedResponseTypes) == 0 {
		allowedResponseTypes = []string{"code"}
	}

	ok := false
	for _, respType := range allowedResponseTypes {
		if respType == r.ResponseType() {
			ok = true
			break
		}
	}
	if !ok {
		return protocol.NewError("unauthorized_client", "response type is not allowed for this client")
	}

	if len(r.Scope()) == 0 {
		return protocol.NewError("invalid_request", "scope is required")
	}

	if slice.ContainsString(r.Prompt(), "none") {
		if len(r.Prompt()) != 1 {
			return protocol.NewError("invalid_request", "prompt cannot have other values when none is set")
		}
		if r.HasMaxAge() {
			return protocol.NewError("invalid_request", "max_age could imply prompt=login so max_age cannot be present when prompt=none")
		}
	}

	switch r.ResponseType() {
	case "code":
		if client.IsPublic() {
			if r.CodeChallenge() == "" {
				return protocol.NewError("invalid_request", "PKCE code challenge is required for public clients")
			}
		}
		if r.CodeChallenge() != "" && r.CodeChallengeMethod() != "S256" {
			return protocol.NewError("invalid_request", "only 'S256' PKCE transform is supported")
		}
	case "none":
		break
	default:
		return protocol.NewError("unsupported_response_type", "only 'code' response type is supported")
	}

	if r.SSOEnabled() && client != nil && client.MaxConcurrentSession == 1 {
		return protocol.NewError("invalid_request", "'sso_enabled' must be false if config 'x_max_concurrent_session' is 1")
	}

	return nil
}

func (h *AuthorizationHandler) generateCodeResponse(
	redirectURI string,
	idpSessionID string,
	authenticationInfo authenticationinfo.T,
	idTokenHintSID string,
	r protocol.AuthorizationRequest,
	authz *oauth.Authorization,
	resp protocol.AuthorizationResponse,
) error {
	code, _, err := h.CodeGrantService.CreateCodeGrant(&CreateCodeGrantOptions{
		Authorization:      authz,
		IDPSessionID:       idpSessionID,
		AuthenticationInfo: authenticationInfo,
		IDTokenHintSID:     idTokenHintSID,
		Scopes:             r.Scope(),
		RedirectURI:        redirectURI,
		OIDCNonce:          r.Nonce(),
		PKCEChallenge:      r.CodeChallenge(),
		SSOEnabled:         r.SSOEnabled(),
	})
	if err != nil {
		return err
	}

	resp.Code(code)
	return nil
}
