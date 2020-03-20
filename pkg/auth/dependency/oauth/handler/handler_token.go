package handler

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	gotime "time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type IDTokenIssuer interface {
	IssueIDToken(client config.OAuthClientConfiguration, userID string, nonce string) (token string, err error)
}

type TokenHandler struct {
	Context context.Context
	Clients []config.OAuthClientConfiguration
	Logger  *logrus.Entry

	Authorizations oauth.AuthorizationStore
	CodeGrants     oauth.CodeGrantStore
	OfflineGrants  oauth.OfflineGrantStore
	AccessGrants   oauth.AccessGrantStore
	Sessions       session.Provider
	IDTokenIssuer  IDTokenIssuer
	GenerateToken  TokenGenerator
	Time           time.Provider
}

func (h *TokenHandler) Handle(r protocol.TokenRequest) TokenResult {
	client, errResp := h.resolveClient(r)
	if errResp != nil {
		return tokenResultError{Response: errResp}
	}

	result, err := h.doHandle(client, r)
	if err != nil {
		var oauthError *protocol.OAuthProtocolError
		resultErr := tokenResultError{}
		if errors.As(err, &oauthError) {
			resultErr.Response = oauthError.Response
		} else {
			h.Logger.WithError(err).Error("authz handler failed")
			resultErr.Response = protocol.NewErrorResponse("server_error", "internal server error")
			resultErr.InternalError = true
		}
		result = resultErr
	}

	return result
}

var errInvalidAuthzCode = protocol.NewError("invalid_grant", "invalid authorization code")

func (h *TokenHandler) doHandle(
	client config.OAuthClientConfiguration,
	r protocol.TokenRequest,
) (TokenResult, error) {
	if err := h.validateRequest(r); err != nil {
		return nil, err
	}

	codeHash := hashToken(r.Code())
	codeGrant, err := h.CodeGrants.GetCodeGrant(codeHash)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, errInvalidAuthzCode
	} else if err != nil {
		return nil, err
	}

	if h.Time.NowUTC().After(codeGrant.ExpireAt) {
		return nil, errInvalidAuthzCode
	}

	if codeGrant.RedirectURI != r.RedirectURI() {
		return nil, protocol.NewError("invalid_request", "invalid redirect URI")
	}

	if codeGrant.PKCEChallenge != "" && !verifyPKCE(codeGrant.PKCEChallenge, r.CodeVerifier()) {
		return nil, errInvalidAuthzCode
	}

	authz, err := h.Authorizations.GetByID(codeGrant.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, errInvalidAuthzCode
	} else if err != nil {
		return nil, err
	}

	sess, err := h.Sessions.Get(codeGrant.SessionID)
	if errors.Is(err, session.ErrSessionNotFound) {
		return nil, errInvalidAuthzCode
	} else if err != nil {
		return nil, err
	}

	resp, err := h.issueTokens(client, codeGrant, authz, sess)
	if err != nil {
		return nil, err
	}

	err = h.CodeGrants.DeleteCodeGrant(codeGrant)
	if err != nil {
		h.Logger.WithError(err).Error("failed to invalidate code grant")
	}

	return tokenResultOK{Response: resp}, nil
}

func (h *TokenHandler) resolveClient(r protocol.TokenRequest) (config.OAuthClientConfiguration, protocol.ErrorResponse) {
	var client config.OAuthClientConfiguration
	for _, c := range h.Clients {
		if c.ClientID() == r.ClientID() {
			client = c
			break
		}
	}
	if client == nil {
		return nil, protocol.NewErrorResponse("invalid_client", "invalid client ID")
	}

	allowedURIs := client.RedirectURIs()
	redirectURIString := r.RedirectURI()
	if len(allowedURIs) == 1 && redirectURIString == "" {
		// Redirect URI is default to the only allowed URI if possible.
		redirectURIString = allowedURIs[0]
	}

	allowed := false
	for _, u := range allowedURIs {
		if u == redirectURIString {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, protocol.NewErrorResponse("invalid_request", "redirect URI is not allowed")
	}

	return client, nil
}

func (h *TokenHandler) validateRequest(r protocol.TokenRequest) error {
	if r.GrantType() != "authorization_code" {
		return protocol.NewError("unsupported_grant_type", "only 'authorization_code' grant type is supported")
	}
	if r.Code() == "" {
		return protocol.NewError("invalid_request", "code is required")
	}
	if r.CodeVerifier() == "" {
		return protocol.NewError("invalid_request", "PKCE code verifier is required")
	}

	return nil
}

func (h *TokenHandler) issueTokens(
	client config.OAuthClientConfiguration,
	code *oauth.CodeGrant,
	authz *oauth.Authorization,
	session *session.IDPSession,
) (protocol.TokenResponse, error) {
	issueRefreshToken := false
	issueIDToken := false
	for _, scope := range code.Scopes {
		switch scope {
		case "offline_access":
			issueRefreshToken = true
		case "openid":
			issueIDToken = true
		}
	}

	resp := protocol.TokenResponse{}

	if issueIDToken {
		if h.IDTokenIssuer == nil {
			return nil, errors.New("id token issuer is not provided")
		}
		idToken, err := h.IDTokenIssuer.IssueIDToken(client, authz.UserID, code.OIDCNonce)
		if err != nil {
			return nil, err
		}
		resp.IDToken(idToken)
	}

	var sessionID string
	var sessionKind oauth.GrantSessionKind
	if issueRefreshToken {
		offlineGrant, err := h.issueOfflineGrant(client, code, authz.ID, session, resp)
		if err != nil {
			return nil, err
		}
		sessionID = offlineGrant.ID
		sessionKind = oauth.GrantSessionKindOffline
	} else {
		sessionID = session.ID
		sessionKind = oauth.GrantSessionKindSession
	}

	err := h.issueAccessGrant(client, code, authz.ID, sessionID, sessionKind, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *TokenHandler) issueOfflineGrant(
	client config.OAuthClientConfiguration,
	code *oauth.CodeGrant,
	authzID string,
	session *session.IDPSession,
	resp protocol.TokenResponse,
) (*oauth.OfflineGrant, error) {
	token := h.GenerateToken()
	now := h.Time.NowUTC()
	offlineGrant := &oauth.OfflineGrant{
		AppID:           code.AppID,
		ID:              uuid.New(),
		AuthorizationID: authzID,

		CreatedAt: now,
		ExpireAt:  now,
		Scopes:    code.Scopes,
		TokenHash: hashToken(token),

		AccessedAt:    now,
		Attrs:         session.Attrs,
		InitialAccess: session.LastAccess,
		LastAccess:    session.LastAccess,
	}
	err := h.OfflineGrants.CreateOfflineGrant(offlineGrant)
	if err != nil {
		return nil, err
	}

	resp.RefreshToken(oauth.EncodeRefreshToken(token, offlineGrant.ID))
	return offlineGrant, nil
}

func (h *TokenHandler) issueAccessGrant(
	client config.OAuthClientConfiguration,
	code *oauth.CodeGrant,
	authzID string,
	sessionID string,
	sessionKind oauth.GrantSessionKind,
	resp protocol.TokenResponse,
) error {
	token := h.GenerateToken()
	now := h.Time.NowUTC()

	accessGrant := &oauth.AccessGrant{
		AppID:           code.AppID,
		AuthorizationID: authzID,
		SessionID:       sessionID,
		SessionKind:     sessionKind,
		CreatedAt:       now,
		ExpireAt:        now.Add(gotime.Duration(client.AccessTokenLifetime()) * gotime.Second),
		Scopes:          code.Scopes,
		TokenHash:       hashToken(token),
	}
	err := h.AccessGrants.CreateAccessGrant(accessGrant)
	if err != nil {
		return err
	}

	resp.TokenType("Bearer")
	resp.AccessToken(oauth.EncodeAccessToken(token))
	resp.ExpiresIn(client.AccessTokenLifetime())
	return nil
}

func verifyPKCE(challenge, verifier string) bool {
	verifierHash := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(verifierHash[:])
	return subtle.ConstantTimeCompare([]byte(challenge), []byte(expectedChallenge)) == 1
}
