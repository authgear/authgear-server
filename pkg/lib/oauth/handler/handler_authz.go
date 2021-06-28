package handler

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
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

	Authorizations  oauth.AuthorizationStore
	CodeGrants      oauth.CodeGrantStore
	OAuthURLs       OAuthURLProvider
	WebAppURLs      WebAppAuthenticateURLProvider
	ValidateScopes  ScopesValidator
	CodeGenerator   TokenGenerator
	LoginHintParser *LoginHintResolver
	IDTokens        IDTokenVerifier
	Clock           clock.Clock
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
	s := session.GetSession(h.Context)
	if s != nil && s.SessionType() != session.TypeIdentityProvider {
		s = nil
	}

	authnOptions := webapp.AuthenticateURLOptions{}

	// Handle max_age and prompt=login
	prompt := r.Prompt()
	if maxAge, ok := r.MaxAge(); ok {
		impliesPromptLogin := false
		// When there is no session, the presence of max_age implies prompt=login.
		if s == nil {
			impliesPromptLogin = true
		} else {
			// max_age=0 implies prompt=login
			if maxAge == 0 {
				impliesPromptLogin = true
			} else {
				// max_age=n implies prompt=login if elapsed time is greater than max_age.
				// In extreme rare case, elapsed time can be negative.
				elapsedTime := h.Clock.NowUTC().Sub(s.GetAuthenticatedAt())
				if elapsedTime < 0 || elapsedTime > maxAge {
					impliesPromptLogin = true
				}
			}
		}
		if impliesPromptLogin {
			prompt = slice.AppendIfUniqueStrings(prompt, "login")
		}
	}
	authnOptions.Prompt = prompt

	var idToken jwt.Token
	// Handle id_token_hint
	if idTokenHint, ok := r.IDTokenHint(); ok {
		idToken, err = h.IDTokens.VerifyIDTokenHint(client, idTokenHint)
		if err != nil {
			return nil, err
		}
		authnOptions.UserIDHint = idToken.Subject()
	}

	// start web app authentication
	if !slice.ContainsString(prompt, "none") {
		r = r.CopyForSelfRedirection()

		// Not authenticated as IdP session => request authentication and retry
		authnOptions.ClientID = r.ClientID()
		authnOptions.UILocales = strings.Join(r.UILocales(), " ")
		authnOptions.Page = r.Page()

		hint, err := h.LoginHintParser.ResolveLoginHint(r.LoginHint())
		if err != nil {
			return nil, protocol.NewError("invalid_request", err.Error())
		}
		authnOptions.AuthenticateHint = hint
		r.SetLoginHint("")

		authorizeURI := h.OAuthURLs.AuthorizeURL(r)
		authnOptions.RedirectURI = authorizeURI.String()
		authnOptions.WebhookState = r.State()

		resp, err := h.WebAppURLs.AuthenticateURL(authnOptions)
		if apierrors.IsKind(err, interaction.InvalidCredentials) {
			return nil, protocol.NewError("invalid_request", err.Error())
		} else if err != nil {
			return nil, err
		}

		return resp, nil
	}

	// start handle prompt == none
	if s == nil || (idToken != nil && s.GetUserID() != idToken.Subject()) {
		return nil, protocol.NewError("login_required", "authentication required")
	}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		r.ClientID(),
		s.GetUserID(),
		scopes,
	)
	if err != nil {
		return nil, err
	}

	resp := protocol.AuthorizationResponse{}
	switch r.ResponseType() {
	case "code":
		err = h.generateCodeResponse(redirectURI.String(), s, r, authz, scopes, resp)
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
	session session.Session,
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
		SessionID:       session.SessionID(),

		CreatedAt: h.Clock.NowUTC(),
		ExpireAt:  h.Clock.NowUTC().Add(CodeGrantValidDuration),
		Scopes:    scopes,
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
