package handler

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type AuthorizationHandler struct {
	Context context.Context
	AppID   string
	Clients []config.OAuthClientConfiguration

	Authorizations       oauth.AuthorizationStore
	CodeGrants           oauth.CodeGrantStore
	AuthorizeEndpoint    AuthorizeEndpointProvider
	AuthenticateEndpoint AuthenticateEndpointProvider
	Time                 time.Provider
}

func (h *AuthorizationHandler) Handle(r protocol.AuthorizationRequest) AuthorizationResult {
	redirectURI, client, errResp := h.resolveClient(r)
	if errResp != nil {
		return authorizationResultError{Response: errResp}
	}

	result, err := h.doHandle(redirectURI, client, r)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		var resp protocol.ErrorResponse
		if errors.As(err, &oauthError) {
			resp = oauthError.Response
		} else {
			resp = protocol.NewErrorResponse("server_error", "internal server error")
		}
		resp.State(r.State())
		result = authorizationResultError{RedirectURI: redirectURI, Response: resp}
	}

	return result
}

func (h *AuthorizationHandler) doHandle(
	redirectURI *url.URL,
	client config.OAuthClientConfiguration,
	r protocol.AuthorizationRequest,
) (AuthorizationResult, error) {
	if err := h.validateRequest(r); err != nil {
		return nil, err
	}

	scopes, err := h.parseScopes(r.Scope())
	if err != nil {
		return nil, err
	}

	session := auth.GetSession(h.Context)
	if session == nil || session.SessionType() != auth.SessionTypeIdentityProvider {
		// Not authenticated as IdP session => request authentication and retry
		return authorizationResultRequireAuthn{
			AuthenticateURI: h.AuthenticateEndpoint.AuthenticateEndpointURI(),
			AuthorizeURI:    h.AuthorizeEndpoint.AuthorizeEndpointURI(),
			Request:         r,
		}, nil
	}

	authz, err := h.checkAuthorization(session, r, scopes)
	if err != nil {
		return nil, err
	}

	code := generateToken()
	codeHash := hashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:           h.AppID,
		AuthorizationID: authz.ID,
		SessionID:       session.SessionID(),

		CreatedAt: h.Time.NowUTC(),
		ExpireAt:  h.Time.NowUTC(),
		Scopes:    scopes,
		CodeHash:  codeHash,

		RedirectURI:   redirectURI.String(),
		OIDCNonce:     r.Nonce(),
		PKCEChallenge: r.CodeChallenge(),
	}

	err = h.CodeGrants.CreateCodeGrant(codeGrant)
	if err != nil {
		return nil, err
	}

	resp := protocol.AuthorizationResponse{}
	resp.Code(code)
	resp.State(r.State())

	return authorizationResultRedirect{
		RedirectURI: redirectURI,
		Response:    resp,
	}, nil
}

func (h *AuthorizationHandler) resolveClient(
	r protocol.AuthorizationRequest,
) (*url.URL, config.OAuthClientConfiguration, protocol.ErrorResponse) {
	var client config.OAuthClientConfiguration
	for _, c := range h.Clients {
		if c.ClientID() == r.ClientID() {
			client = c
			break
		}
	}
	if client == nil {
		return nil, nil, protocol.NewErrorResponse("unauthorized_client", "invalid client ID")
	}

	allowedURIs := client.RedirectURIs()
	redirectURIString := r.RedirectURI()
	if len(allowedURIs) == 1 && redirectURIString == "" {
		// Redirect URI is default to the only allowed URI if possible.
		redirectURIString = allowedURIs[0]
	}

	redirectURI, err := url.Parse(redirectURIString)
	if err != nil {
		return nil, nil, protocol.NewErrorResponse("invalid_request", "invalid redirect URI")
	}

	allowed := false
	for _, u := range allowedURIs {
		if u == redirectURIString {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, nil, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed")
	}

	return redirectURI, client, nil
}

func (h *AuthorizationHandler) validateRequest(r protocol.AuthorizationRequest) error {
	if r.ResponseType() != "code" {
		return protocol.NewError("unsupported_response_type", "only 'code' response type is supported")
	}
	if r.Scope() == "" {
		return protocol.NewError("invalid_request", "scope is required")
	}
	if r.CodeChallengeMethod() != "S256" {
		return protocol.NewError("invalid_request", "only 'S256' PKCE transform is supported")
	}
	if r.CodeChallenge() == "" {
		return protocol.NewError("invalid_request", "PKCE code challenge is required")
	}

	return nil
}

func (h *AuthorizationHandler) parseScopes(scope string) ([]string, error) {
	scopes := strings.Split(scope, " ")
	if err := oidc.ValidateScopes(scopes); err != nil {
		return nil, err
	}
	return scopes, nil
}

func (h *AuthorizationHandler) checkAuthorization(
	session auth.AuthSession,
	r protocol.AuthorizationRequest,
	scopes []string,
) (*oauth.Authorization, error) {
	userID := session.AuthnAttrs().UserID
	authz, err := h.Authorizations.Get(userID, r.ClientID())
	if err == nil && authz.IsAuthorized(scopes) {
		return authz, nil
	} else if err != nil && !errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, err
	}

	// Authorization of requested scopes not granted, requesting consent.
	// TODO(oauth): request consent, for now just always implicitly grant scopes.
	if authz == nil {
		now := h.Time.NowUTC()
		authz = &oauth.Authorization{
			ID:        uuid.New(),
			AppID:     h.AppID,
			ClientID:  r.ClientID(),
			UserID:    userID,
			CreatedAt: now,
			UpdatedAt: now,
			Scopes:    scopes,
		}
		err = h.Authorizations.Create(authz)
		if err != nil {
			return nil, err
		}
	} else {
		authz = authz.WithScopesAdded(scopes)
		err = h.Authorizations.UpdateScopes(authz)
		if err != nil {
			return nil, err
		}
	}

	return authz, nil
}
