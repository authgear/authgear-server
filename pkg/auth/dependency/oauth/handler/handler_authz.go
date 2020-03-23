package handler

import (
	"context"
	"errors"
	"net/url"
	gotime "time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

const CodeGrantValidDuration = 5 * gotime.Minute

type AuthorizationHandler struct {
	Context context.Context
	AppID   string
	Clients []config.OAuthClientConfiguration
	Logger  *logrus.Entry

	Authorizations       oauth.AuthorizationStore
	CodeGrants           oauth.CodeGrantStore
	AuthorizeEndpoint    oauth.AuthorizeEndpointProvider
	AuthenticateEndpoint oauth.AuthenticateEndpointProvider
	ValidateScopes       ScopesValidator
	CodeGenerator        TokenGenerator
	Time                 time.Provider
}

func (h *AuthorizationHandler) Handle(r protocol.AuthorizationRequest) AuthorizationResult {
	client := resolveClient(h.Clients, r)
	if client == nil {
		return authorizationResultError{
			Response: protocol.NewErrorResponse("unauthorized_client", "invalid client ID"),
		}
	}
	redirectURI, errResp := parseRedirectURI(client, r)
	if errResp != nil {
		return authorizationResultError{Response: errResp}
	}

	result, err := h.doHandle(redirectURI, client, r)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		resultErr := authorizationResultError{RedirectURI: redirectURI}
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
	client config.OAuthClientConfiguration,
	r protocol.AuthorizationRequest,
) (AuthorizationResult, error) {
	if err := h.validateRequest(r); err != nil {
		return nil, err
	}

	scopes := r.Scope()
	err := h.ValidateScopes(client, scopes)
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

	code := h.CodeGenerator()
	codeHash := hashToken(code)

	codeGrant := &oauth.CodeGrant{
		AppID:           h.AppID,
		AuthorizationID: authz.ID,
		SessionID:       session.SessionID(),

		CreatedAt: h.Time.NowUTC(),
		ExpireAt:  h.Time.NowUTC().Add(CodeGrantValidDuration),
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
	state := r.State()
	if state != "" {
		resp.State(r.State())
	}

	return authorizationResultRedirect{
		RedirectURI: redirectURI,
		Response:    resp,
	}, nil
}

func (h *AuthorizationHandler) validateRequest(r protocol.AuthorizationRequest) error {
	if r.ResponseType() != "code" {
		return protocol.NewError("unsupported_response_type", "only 'code' response type is supported")
	}
	if len(r.Scope()) == 0 {
		return protocol.NewError("invalid_request", "scope is required")
	}
	if r.CodeChallenge() == "" {
		return protocol.NewError("invalid_request", "PKCE code challenge is required")
	}
	if r.CodeChallengeMethod() != "S256" {
		return protocol.NewError("invalid_request", "only 'S256' PKCE transform is supported")
	}

	return nil
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
		authz.UpdatedAt = h.Time.NowUTC()
		err = h.Authorizations.UpdateScopes(authz)
		if err != nil {
			return nil, err
		}
	}

	return authz, nil
}
