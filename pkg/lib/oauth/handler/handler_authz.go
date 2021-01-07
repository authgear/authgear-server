package handler

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

const CodeGrantValidDuration = 5 * time.Minute

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

	s := session.GetSession(h.Context)
	authnOptions := webapp.AuthenticateURLOptions{}
	authnOptions.Platform = r.Platform()
	if slice.ContainsString(r.Prompt(), "login") {
		// Request login prompt => force re-authentication and retry
		r2 := protocol.AuthorizationRequest{}
		for k, v := range r {
			r2[k] = v
		}
		prompt := slice.ExceptStrings(r.Prompt(), []string{"login"})
		r2.SetPrompt(prompt)
		authnOptions.Prompt = "login"

		r = r2
		// Treat as not authenticated
		s = nil
	}
	if s == nil || s.SessionType() != session.TypeIdentityProvider {
		// Not authenticated as IdP session => request authentication and retry
		authnOptions.ClientID = r.ClientID()
		authnOptions.UILocales = strings.Join(r.UILocales(), " ")

		hint, err := h.LoginHintParser.ResolveLoginHint(r.LoginHint())
		if err != nil {
			return nil, protocol.NewError("invalid_request", err.Error())
		}
		authnOptions.AuthenticateHint = hint
		r.SetLoginHint("")

		authorizeURI := h.OAuthURLs.AuthorizeURL(r)
		authnOptions.RedirectURI = authorizeURI.String()

		resp, err := h.WebAppURLs.AuthenticateURL(authnOptions)
		if apierrors.IsKind(err, interaction.InvalidCredentials) {
			return nil, protocol.NewError("invalid_request", err.Error())
		} else if err != nil {
			return nil, err
		}

		return resp, nil
	}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		r.ClientID(),
		s.SessionAttrs().UserID,
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

	if slice.ContainsString(r.Prompt(), "none") && len(r.Prompt()) != 1 {
		return protocol.NewError("invalid_request", "prompt cannot have other values when none is set")
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
