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
	Consume(entryID string) (*authenticationinfo.Entry, error)
}

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type SessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
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

func (h *AuthorizationHandler) HandleFromWebApp(r protocol.AuthorizationRequest, req *http.Request) httputil.Result {
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

	result, err := h.doHandleFromWebApp(redirectURI, client, r, req)
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

	// Handle login_hint
	// We must return here.
	if loginHint, ok := r.LoginHint(); ok {
		result, err := h.LoginHint.HandleLoginHint(webapp.HandleLoginHintOptions{
			SessionOptions:      sessionOptions,
			LoginHint:           loginHint,
			UILocales:           uiLocales,
			ColorScheme:         colorScheme,
			OriginalRedirectURI: redirectURI.String(),
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

	return h.finish(redirectURI, r, idpSession.SessionID(), authenticationInfo, idTokenHintSID)
}

func (h *AuthorizationHandler) finish(
	redirectURI *url.URL,
	r protocol.AuthorizationRequest,
	idpSessionID string,
	authenticationInfo authenticationinfo.T,
	idTokenHintSID string,
) (httputil.Result, error) {
	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		r.ClientID(),
		authenticationInfo.UserID,
		r.Scope(),
	)
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
		Cookies:      []*http.Cookie{h.Cookies.ClearCookie(authenticationinfo.CookieDef)},
	}, nil
}

func (h *AuthorizationHandler) doHandleFromWebApp(
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
	req *http.Request,
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

	cookie, err := h.Cookies.GetCookie(req, authenticationinfo.CookieDef)
	if err != nil {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	entry, err := h.AuthenticationInfoService.Consume(cookie.Value)
	if err != nil {
		return nil, err
	}

	authenticationInfo := entry.T
	return h.finish(redirectURI, r, idpSessionID, authenticationInfo, idTokenHintSID)
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
