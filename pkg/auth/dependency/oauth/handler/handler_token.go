package handler

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	gotime "time"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

// TODO(oauth): write tests

type IDTokenIssuer interface {
	IssueIDToken(client config.OAuthClientConfiguration, session auth.AuthSession, nonce string) (token string, err error)
}

type TokenHandler struct {
	Request *http.Request
	Clients []config.OAuthClientConfiguration
	Logger  *logrus.Entry

	Authorizations oauth.AuthorizationStore
	CodeGrants     oauth.CodeGrantStore
	OfflineGrants  oauth.OfflineGrantStore
	AccessGrants   oauth.AccessGrantStore
	AccessEvents   auth.AccessEventProvider
	Sessions       session.Provider
	IDTokenIssuer  IDTokenIssuer
	GenerateToken  TokenGenerator
	Time           time.Provider
}

func (h *TokenHandler) Handle(r protocol.TokenRequest) TokenResult {
	client := resolveClient(h.Clients, r)
	if client == nil {
		return authorizationResultError{
			Response: protocol.NewErrorResponse("invalid_client", "invalid client ID"),
		}
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

func (h *TokenHandler) doHandle(
	client config.OAuthClientConfiguration,
	r protocol.TokenRequest,
) (TokenResult, error) {
	if err := h.validateRequest(r); err != nil {
		return nil, err
	}

	allowedGrantTypes := client.GrantTypes()
	if len(allowedGrantTypes) == 0 {
		allowedGrantTypes = []string{"authorization_code"}
	}
	ok := false
	for _, grantType := range allowedGrantTypes {
		if r.GrantType() == grantType {
			ok = true
			break
		}
	}
	if !ok {
		return nil, protocol.NewError("unauthorized_client", "grant type is not allowed for this client")
	}

	switch r.GrantType() {
	case "authorization_code":
		return h.handleAuthorizationCode(client, r)
	case "refresh_token":
		return h.handleRefreshToken(client, r)
	default:
		panic("oauth: unexpected grant type")
	}
}

func (h *TokenHandler) validateRequest(r protocol.TokenRequest) error {
	switch r.GrantType() {
	case "authorization_code":
		if r.Code() == "" {
			return protocol.NewError("invalid_request", "code is required")
		}
		if r.CodeVerifier() == "" {
			return protocol.NewError("invalid_request", "PKCE code verifier is required")
		}
	case "refresh_token":
		if r.RefreshToken() == "" {
			return protocol.NewError("invalid_request", "refresh token is required")
		}
	default:
		return protocol.NewError("unsupported_grant_type",
			"only 'authorization_code' and 'refresh_token' grant types are supported")
	}

	return nil
}

var errInvalidAuthzCode = protocol.NewError("invalid_grant", "invalid authorization code")

func (h *TokenHandler) handleAuthorizationCode(
	client config.OAuthClientConfiguration,
	r protocol.TokenRequest,
) (TokenResult, error) {

	codeHash := oauth.HashToken(r.Code())
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

	resp, err := h.issueTokensForAuthorizationCode(client, codeGrant, authz, sess)
	if err != nil {
		return nil, err
	}

	err = h.CodeGrants.DeleteCodeGrant(codeGrant)
	if err != nil {
		h.Logger.WithError(err).Error("failed to invalidate code grant")
	}

	return tokenResultOK{Response: resp}, nil
}

var errInvalidRefreshToken = protocol.NewError("invalid_grant", "invalid refresh token")

func (h *TokenHandler) handleRefreshToken(
	client config.OAuthClientConfiguration,
	r protocol.TokenRequest,
) (TokenResult, error) {
	token, grantID, err := oauth.DecodeRefreshToken(r.RefreshToken())
	if err != nil {
		return nil, errInvalidRefreshToken
	}

	offlineGrant, err := h.OfflineGrants.GetOfflineGrant(grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, err
	}

	if h.Time.NowUTC().After(offlineGrant.ExpireAt) {
		return nil, errInvalidRefreshToken
	}

	tokenHash := oauth.HashToken(token)
	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(offlineGrant.TokenHash)) != 1 {
		return nil, errInvalidRefreshToken
	}

	authz, err := h.Authorizations.GetByID(offlineGrant.AuthorizationID)
	if errors.Is(err, oauth.ErrAuthorizationNotFound) {
		return nil, errInvalidRefreshToken
	} else if err != nil {
		return nil, err
	}

	resp, err := h.issueTokensForRefreshToken(client, offlineGrant, authz)
	if err != nil {
		return nil, err
	}

	return tokenResultOK{Response: resp}, nil
}

func (h *TokenHandler) issueTokensForAuthorizationCode(
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

	if issueRefreshToken {
		// Only if client is allowed to use refresh tokens
		allowRefreshToken := false
		for _, grantType := range client.GrantTypes() {
			if grantType == "refresh_token" {
				allowRefreshToken = true
				break
			}
		}
		if !allowRefreshToken {
			issueRefreshToken = false
		}
	}

	resp := protocol.TokenResponse{}

	var sessionID string
	var sessionKind oauth.GrantSessionKind
	var atSession auth.AuthSession
	if issueRefreshToken {
		offlineGrant, err := h.issueOfflineGrant(client, code, authz.ID, session.AuthnAttrs(), resp)
		if err != nil {
			return nil, err
		}
		atSession = offlineGrant
		sessionID = offlineGrant.ID
		sessionKind = oauth.GrantSessionKindOffline
	} else {
		atSession = session
		sessionID = session.ID
		sessionKind = oauth.GrantSessionKindSession
	}

	err := h.issueAccessGrant(client, code.AppID, code.Scopes,
		authz.ID, sessionID, sessionKind, resp)
	if err != nil {
		return nil, err
	}

	if issueIDToken {
		if h.IDTokenIssuer == nil {
			return nil, errors.New("id token issuer is not provided")
		}
		idToken, err := h.IDTokenIssuer.IssueIDToken(client, atSession, code.OIDCNonce)
		if err != nil {
			return nil, err
		}
		resp.IDToken(idToken)
	}

	return resp, nil
}

func (h *TokenHandler) issueTokensForRefreshToken(
	client config.OAuthClientConfiguration,
	offlineGrant *oauth.OfflineGrant,
	authz *oauth.Authorization,
) (protocol.TokenResponse, error) {
	issueIDToken := false
	for _, scope := range offlineGrant.Scopes {
		if scope == "openid" {
			issueIDToken = true
			break
		}
	}

	resp := protocol.TokenResponse{}

	if issueIDToken {
		if h.IDTokenIssuer == nil {
			return nil, errors.New("id token issuer is not provided")
		}
		idToken, err := h.IDTokenIssuer.IssueIDToken(client, offlineGrant, "")
		if err != nil {
			return nil, err
		}
		resp.IDToken(idToken)
	}

	err := h.issueAccessGrant(client, offlineGrant.AppID, offlineGrant.Scopes,
		authz.ID, offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *TokenHandler) issueOfflineGrant(
	client config.OAuthClientConfiguration,
	code *oauth.CodeGrant,
	authzID string,
	attrs *authn.Attrs,
	resp protocol.TokenResponse,
) (*oauth.OfflineGrant, error) {
	token := h.GenerateToken()
	now := h.Time.NowUTC()
	accessEvent := auth.NewAccessEvent(now, h.Request)
	offlineGrant := &oauth.OfflineGrant{
		AppID:           code.AppID,
		ID:              uuid.New(),
		AuthorizationID: authzID,
		ClientID:        client.ClientID(),

		CreatedAt: now,
		ExpireAt:  now.Add(gotime.Duration(client.RefreshTokenLifetime()) * gotime.Second),
		Scopes:    code.Scopes,
		TokenHash: oauth.HashToken(token),

		Attrs: *attrs,
		AccessInfo: auth.AccessInfo{
			InitialAccess: accessEvent,
			LastAccess:    accessEvent,
		},
	}
	err := h.OfflineGrants.CreateOfflineGrant(offlineGrant)
	if err != nil {
		return nil, err
	}

	err = h.AccessEvents.InitStream(offlineGrant)
	if err != nil {
		return nil, err
	}

	resp.RefreshToken(oauth.EncodeRefreshToken(token, offlineGrant.ID))
	return offlineGrant, nil
}

func (h *TokenHandler) issueAccessGrant(
	client config.OAuthClientConfiguration,
	appID string,
	scopes []string,
	authzID string,
	sessionID string,
	sessionKind oauth.GrantSessionKind,
	resp protocol.TokenResponse,
) error {
	token := h.GenerateToken()
	now := h.Time.NowUTC()

	accessGrant := &oauth.AccessGrant{
		AppID:           appID,
		AuthorizationID: authzID,
		SessionID:       sessionID,
		SessionKind:     sessionKind,
		CreatedAt:       now,
		ExpireAt:        now.Add(gotime.Duration(client.AccessTokenLifetime()) * gotime.Second),
		Scopes:          scopes,
		TokenHash:       oauth.HashToken(token),
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
