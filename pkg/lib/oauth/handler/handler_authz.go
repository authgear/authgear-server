package handler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
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
	AuthorizeURL(r protocol.AuthorizationRequest) *url.URL
}

type WebAppAuthenticateURLProvider interface {
	AuthenticateURL(options webapp.AuthenticateURLOptions) (httputil.Result, error)
}

type LoginHintHandler interface {
	HandleLoginHint(options webapp.HandleLoginHintOptions) (httputil.Result, error)
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

	Sessions       SessionProvider
	Authorizations oauth.AuthorizationStore
	OfflineGrants  oauth.OfflineGrantStore
	CodeGrants     oauth.CodeGrantStore
	OAuthURLs      OAuthURLProvider
	WebAppURLs     WebAppAuthenticateURLProvider
	ValidateScopes ScopesValidator
	CodeGenerator  TokenGenerator
	LoginHint      LoginHintHandler
	IDTokens       IDTokenVerifier
	Clock          clock.Clock
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

// nolint: gocyclo
func (h *AuthorizationHandler) doHandle(
	redirectURI *url.URL,
	client *config.OAuthClientConfig,
	r protocol.AuthorizationRequest,
) (httputil.Result, error) {
	if err := h.validateRequest(client, r); err != nil {
		return nil, err
	}

	scopes := r.Scope()
	err := h.ValidateScopes(client, scopes)
	if err != nil {
		return nil, err
	}

	// Authorization endpoint ignores non-IDP session.
	var idpSession session.Session
	if s := session.GetSession(h.Context); s != nil && s.SessionType() == session.TypeIdentityProvider {
		idpSession = s
	}

	sessionOptions := webapp.SessionOptions{
		ClientID:     r.ClientID(),
		WebhookState: r.State(),
		Page:         r.Page(),
	}
	uiLocales := strings.Join(r.UILocales(), " ")

	// Handle max_age and prompt=login
	prompt := r.Prompt()
	if maxAge, ok := r.MaxAge(); ok {
		impliesPromptLogin := false
		// When there is no session, the presence of max_age implies prompt=login.
		if idpSession == nil {
			impliesPromptLogin = true
		} else {
			// max_age=0 implies prompt=login
			if maxAge == 0 {
				impliesPromptLogin = true
			} else {
				// max_age=n implies prompt=login if elapsed time is greater than max_age.
				// In extreme rare case, elapsed time can be negative.
				elapsedTime := h.Clock.NowUTC().Sub(idpSession.GetAuthenticatedAt())
				if elapsedTime < 0 || elapsedTime > maxAge {
					impliesPromptLogin = true
				}
			}
		}
		if impliesPromptLogin {
			prompt = slice.AppendIfUniqueStrings(prompt, "login")
		}
	}
	sessionOptions.Prompt = prompt

	// Handle id_token_hint
	var idToken jwt.Token
	if idTokenHint, ok := r.IDTokenHint(); ok {
		idToken, err = h.IDTokens.VerifyIDTokenHint(client, idTokenHint)
		if err != nil {
			return nil, err
		}
		sessionOptions.UserIDHint = idToken.Subject()

		// Set CanUseIntentReauthenticate to true if
		// 1. The ID token is not expired.
		// 2. The sid of the ID token hint points to a valid session (either IDPSession or OfflineGrant).
		// 3. If the session indicated by sid is an offline grant, it must contain the FullAccessScope.
		if tv := idToken.Expiration(); !tv.IsZero() && tv.Unix() != 0 {
			now := h.Clock.NowUTC().Truncate(time.Second)
			tv = tv.Truncate(time.Second)
			if now.Before(tv) {
				if sid, ok := idToken.Get(string(authn.ClaimSID)); ok {
					if sid, ok := sid.(string); ok {
						if typ, sessionID, ok := oidc.DecodeSID(sid); ok {
							switch typ {
							case session.TypeIdentityProvider:
								if _, err = h.Sessions.Get(sessionID); err == nil {
									sessionOptions.CanUseIntentReauthenticate = true
								}
							case session.TypeOfflineGrant:
								if offlineGrant, err := h.OfflineGrants.GetOfflineGrant(sessionID); err == nil {
									if slice.ContainsString(offlineGrant.Scopes, oauth.FullAccessScope) {
										sessionOptions.CanUseIntentReauthenticate = true
									}
								}
							default:
								panic(fmt.Errorf("oauth: unknown session type: %v", typ))
							}
						}
					}
				}
			}
		}

	}

	loginHint, hasLoginHint := r.PopLoginHint()

	// Generate self redirect URI here.
	// Note that it is important to have Prompt, UserIDHint, and LoginHint processed before
	// the URI is generated here.
	r = r.CopyForSelfRedirection()
	authorizeURI := h.OAuthURLs.AuthorizeURL(r)
	sessionOptions.RedirectURI = authorizeURI.String()

	// Handle login_hint
	// We must return here.
	if hasLoginHint {
		result, err := h.LoginHint.HandleLoginHint(webapp.HandleLoginHintOptions{
			SessionOptions:      sessionOptions,
			LoginHint:           loginHint,
			UILocales:           uiLocales,
			OriginalRedirectURI: redirectURI.String(),
		})
		if err != nil {
			return nil, protocol.NewError("invalid_request", err.Error())
		}
		return result, nil
	}

	// Handle prompt!=none
	if !slice.ContainsString(prompt, "none") {
		resp, err := h.WebAppURLs.AuthenticateURL(webapp.AuthenticateURLOptions{
			SessionOptions: sessionOptions,
			UILocales:      uiLocales,
		})
		if apierrors.IsKind(err, interaction.InvalidCredentials) {
			return nil, protocol.NewError("invalid_request", err.Error())
		} else if err != nil {
			return nil, err
		}

		return resp, nil
	}

	// Handle prompt=none
	if idpSession == nil || (idToken != nil && idpSession.GetUserID() != idToken.Subject()) {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		r.ClientID(),
		idpSession.GetUserID(),
		scopes,
	)
	if err != nil {
		return nil, err
	}

	resp := protocol.AuthorizationResponse{}
	switch r.ResponseType() {
	case "code":
		err = h.generateCodeResponse(redirectURI.String(), idpSession, idToken, r, authz, scopes, resp)
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
	}, nil
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
		if r.CodeChallenge() == "" {
			return protocol.NewError("invalid_request", "PKCE code challenge is required")
		}
		if r.CodeChallengeMethod() != "S256" {
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
	idpSession session.Session,
	idTokenHint jwt.Token,
	r protocol.AuthorizationRequest,
	authz *oauth.Authorization,
	scopes []string,
	resp protocol.AuthorizationResponse,
) error {
	code := h.CodeGenerator()
	codeHash := oauth.HashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:           string(h.AppID),
		AuthorizationID: authz.ID,
		IDPSessionID:    idpSession.SessionID(),

		CreatedAt: h.Clock.NowUTC(),
		ExpireAt:  h.Clock.NowUTC().Add(CodeGrantValidDuration),
		Scopes:    scopes,
		CodeHash:  codeHash,

		RedirectURI:   redirectURI,
		OIDCNonce:     r.Nonce(),
		PKCEChallenge: r.CodeChallenge(),
	}

	if idTokenHint != nil {
		if sid, ok := idTokenHint.Get(string(authn.ClaimSID)); ok {
			if sid, ok := sid.(string); ok {
				codeGrant.IDTokenHintSID = sid
			}
		}
	}

	err := h.CodeGrants.CreateCodeGrant(codeGrant)
	if err != nil {
		return err
	}

	resp.Code(code)
	return nil
}
