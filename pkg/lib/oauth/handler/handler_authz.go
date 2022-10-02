package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

const CodeGrantValidDuration = duration.Short

type IDTokenVerifier interface {
	VerifyIDTokenHint(client *config.OAuthClientConfig, idTokenHint string) (jwt.Token, error)
}

type OAuthURLProvider interface {
	FromWebAppURL(r protocol.AuthorizationRequest) *url.URL
}

type WebAppAuthenticateURLProvider interface {
	AuthenticateURL(options webapp.AuthenticateURLOptions) (httputil.Result, error)
}

type LoginHintHandler interface {
	HandleLoginHint(options webapp.HandleLoginHintOptions) (httputil.Result, error)
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

type SessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
}

type OAuthSessionService interface {
	Save(entry *oauthsession.Entry) (err error)
	Get(entryID string) (*oauthsession.Entry, error)
	Delete(entryID string) error
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

	Sessions                  SessionProvider
	Authorizations            oauth.AuthorizationStore
	OfflineGrants             oauth.OfflineGrantStore
	CodeGrants                oauth.CodeGrantStore
	OAuthURLs                 OAuthURLProvider
	WebAppURLs                WebAppAuthenticateURLProvider
	ValidateScopes            ScopesValidator
	CodeGenerator             TokenGenerator
	LoginHint                 LoginHintHandler
	IDTokens                  IDTokenVerifier
	AuthenticationInfoService AuthenticationInfoService
	Clock                     clock.Clock
	Cookies                   CookieManager
	OAuthSessionService       OAuthSessionService
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
	autoGrantAuthz := client.ClientParty() == config.ClientPartyFirst
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
		return nil, err
	}

	return &consentRequest{
		OAuthSessionEntry: entry,
		AuthInfoEntry:     authInfoEntry,
		RedirectURI:       redirectURI,
		Client:            client,
	}, nil
}

func (h *AuthorizationHandler) HandleFromWebApp(req *http.Request) httputil.Result {
	cookie, err := h.Cookies.GetCookie(req, oauthsession.CookieDef)
	if err != nil {
		return authorizationResultError{
			// failed to obtain authz request, use default response mode and empty redirect uri
			// the error will be rendered on the browser without redirection
			ResponseMode: "",
			Response:     protocol.NewErrorResponse("invalid_request", "missing oauth session cookie"),
			RedirectURI:  nil,
		}
	}

	entry, err := h.OAuthSessionService.Get(cookie.Value)
	if err != nil {
		if errors.Is(err, oauthsession.ErrNotFound) {
			// failed to obtain authz request, use default response mode and empty redirect uri
			// the error will be rendered on the browser without redirection
			return authorizationResultError{
				ResponseMode: "",
				Response:     protocol.NewErrorResponse("invalid_request", "oauth session expired"),
				RedirectURI:  nil,
			}
		}
		h.Logger.WithError(err).Error("failed to obtain oauth session")
		return authorizationResultError{
			// failed to obtain authz request, use default response mode and empty redirect uri
			// the error will be rendered on the browser without redirection
			ResponseMode:  "",
			Response:      protocol.NewErrorResponse("server_error", "internal server error"),
			InternalError: true,
			RedirectURI:   nil,
		}
	}

	r := entry.T.AuthorizationRequest

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

	authInfoEntry, err := h.getAuthenticationInfoEntry(req)
	if err != nil {
		return authorizationResultError{
			ResponseMode: r.ResponseMode(),
			Response:     protocol.NewErrorResponse("login_required", "authentication required"),
		}
	}
	result, err := h.doHandleConsentRequest(redirectURI, client, r, authInfoEntry.T, req, true)
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
	} else {
		// delete oauth session with best effort
		// don't block the user in case of failure
		err = h.OAuthSessionService.Delete(entry.ID)
		if err != nil {
			h.Logger.WithError(err).Error("failed to consume oauth session")
		}
	}

	return result
}

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

	idToken, sidSession, err := h.handleIDTokenHint(client, r)
	if err != nil {
		return nil, err
	}

	var idTokenHintSID string
	if sidSession != nil {
		idTokenHintSID = oidc.EncodeSID(sidSession)
	}

	sessionOptions := webapp.SessionOptions{
		ClientID:                 r.ClientID(),
		WebhookState:             r.State(),
		Page:                     r.Page(),
		RedirectURI:              h.OAuthURLs.FromWebAppURL(r).String(),
		Prompt:                   h.handleMaxAgeAndPrompt(r, sidSession),
		SuppressIDPSessionCookie: r.SuppressIDPSessionCookie(),
		OAuthProviderAlias:       r.OAuthProviderAlias(),
	}
	uiLocales := strings.Join(r.UILocales(), " ")
	colorScheme := r.ColorScheme()

	if idToken != nil {
		sessionOptions.UserIDHint = idToken.Subject()
		// Set CanUseIntentReauthenticate to true if
		// 1. The ID token is not expired.
		// 2. The sid of the ID token hint points to a valid session (either IDPSession or OfflineGrant).
		// 3. If the session indicated by sid is an offline grant, it must contain the FullAccessScope.
		if tv := idToken.Expiration(); !tv.IsZero() && tv.Unix() != 0 {
			now := h.Clock.NowUTC().Truncate(time.Second)
			tv = tv.Truncate(time.Second)
			if now.Before(tv) && sidSession != nil {
				switch sidSession.SessionType() {
				case session.TypeIdentityProvider:
					sessionOptions.CanUseIntentReauthenticate = true
				case session.TypeOfflineGrant:
					if offlineGrant, ok := sidSession.(*oauth.OfflineGrant); ok {
						if slice.ContainsString(offlineGrant.Scopes, oauth.FullAccessScope) {
							sessionOptions.CanUseIntentReauthenticate = true
						}
					}
				}
			}
		}
	}

	// create oauth session and redirect to the web app
	oauthSessionEntry := oauthsession.NewEntry(oauthsession.T{
		AuthorizationRequest: r,
	})
	err = h.OAuthSessionService.Save(oauthSessionEntry)
	if err != nil {
		return nil, err
	}
	oauthSessionEntryCookies := []*http.Cookie{
		h.Cookies.ValueCookie(oauthsession.CookieDef, oauthSessionEntry.ID),
	}

	// Handle login_hint
	// We must return here.
	if loginHint, ok := r.LoginHint(); ok {
		result, err := h.LoginHint.HandleLoginHint(webapp.HandleLoginHintOptions{
			SessionOptions:      sessionOptions,
			LoginHint:           loginHint,
			UILocales:           uiLocales,
			ColorScheme:         colorScheme,
			OriginalRedirectURI: redirectURI.String(),
			OAuthSessionCookies: oauthSessionEntryCookies,
		})
		if err != nil {
			return nil, protocol.NewError("invalid_request", err.Error())
		}
		return result, nil
	}

	// Handle prompt!=none
	// We must return here.
	if !slice.ContainsString(sessionOptions.Prompt, "none") {
		resp, err := h.WebAppURLs.AuthenticateURL(webapp.AuthenticateURLOptions{
			SessionOptions: sessionOptions,
			UILocales:      uiLocales,
			ColorScheme:    colorScheme,
			Cookies:        oauthSessionEntryCookies,
		})
		if apierrors.IsKind(err, interaction.InvalidCredentials) {
			return nil, protocol.NewError("invalid_request", err.Error())
		} else if err != nil {
			return nil, err
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
	autoGrantAuthz := client.ClientParty() == config.ClientPartyFirst

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
		authz, err = checkAndGrantAuthorization(
			h.Authorizations,
			h.Clock.NowUTC(),
			h.AppID,
			r.ClientID(),
			authenticationInfo.UserID,
			r.Scope(),
		)
	} else {
		authz, err = checkAuthorization(
			h.Authorizations,
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

	_, sidSession, err := h.handleIDTokenHint(client, r)
	if err != nil {
		return nil, err
	}

	var idTokenHintSID string
	if sidSession != nil {
		idTokenHintSID = oidc.EncodeSID(sidSession)
	}

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
		if client.ClientParty() == config.ClientPartyFirst {
			if r.CodeChallenge() == "" {
				return protocol.NewError("invalid_request", "PKCE code challenge is required")
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
	code := h.CodeGenerator()
	codeHash := oauth.HashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:              string(h.AppID),
		AuthorizationID:    authz.ID,
		IDPSessionID:       idpSessionID,
		AuthenticationInfo: authenticationInfo,
		IDTokenHintSID:     idTokenHintSID,

		CreatedAt: h.Clock.NowUTC(),
		ExpireAt:  h.Clock.NowUTC().Add(CodeGrantValidDuration),
		Scopes:    r.Scope(),
		CodeHash:  codeHash,

		RedirectURI:   redirectURI,
		OIDCNonce:     r.Nonce(),
		PKCEChallenge: r.CodeChallenge(),
	}

	err := h.CodeGrants.CreateCodeGrant(codeGrant)
	if err != nil {
		return err
	}

	resp.Code(code)
	return nil
}

func (h *AuthorizationHandler) handleMaxAgeAndPrompt(
	r protocol.AuthorizationRequest,
	sidSession session.Session,
) (prompt []string) {
	prompt = r.Prompt()
	if maxAge, ok := r.MaxAge(); ok {
		impliesPromptLogin := false
		// When there is no session, the presence of max_age implies prompt=login.
		if sidSession == nil {
			impliesPromptLogin = true
		} else {
			// max_age=0 implies prompt=login
			if maxAge == 0 {
				impliesPromptLogin = true
			} else {
				// max_age=n implies prompt=login if elapsed time is greater than max_age.
				// In extreme rare case, elapsed time can be negative.
				elapsedTime := h.Clock.NowUTC().Sub(sidSession.GetAuthenticationInfo().AuthenticatedAt)
				if elapsedTime < 0 || elapsedTime > maxAge {
					impliesPromptLogin = true
				}
			}
		}
		if impliesPromptLogin {
			prompt = slice.AppendIfUniqueStrings(prompt, "login")
		}
	}

	return
}

func (h *AuthorizationHandler) handleIDTokenHint(
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) (idToken jwt.Token, sidSession session.Session, err error) {
	idTokenHint, ok := r.IDTokenHint()
	if !ok {
		return
	}

	idToken, err = h.IDTokens.VerifyIDTokenHint(client, idTokenHint)
	if err != nil {
		return
	}

	sidInterface, ok := idToken.Get(string(model.ClaimSID))
	if !ok {
		return
	}

	sid, ok := sidInterface.(string)
	if !ok {
		return
	}

	typ, sessionID, ok := oidc.DecodeSID(sid)
	if !ok {
		return
	}

	switch typ {
	case session.TypeIdentityProvider:
		if sess, err := h.Sessions.Get(sessionID); err == nil {
			sidSession = sess
		}
	case session.TypeOfflineGrant:
		if sess, err := h.OfflineGrants.GetOfflineGrant(sessionID); err == nil {
			sidSession = sess
		}
	default:
		panic(fmt.Errorf("oauth: unknown session type: %v", typ))
	}

	return
}
