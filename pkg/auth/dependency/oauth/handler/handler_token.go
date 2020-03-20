package handler

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type TokenHandler struct {
	Context context.Context
	Clients []config.OAuthClientConfiguration
	Logger  *logrus.Entry

	Authorizations oauth.AuthorizationStore
	CodeGrants     oauth.CodeGrantStore
	OfflineGrants  oauth.OfflineGrantStore
	AccessGrants   oauth.AccessGrantStore
	Sessions       session.Provider
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

	resp, err := h.issueTokens(codeGrant, authz, sess)
	if err != nil {
		return nil, err
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
	code *oauth.CodeGrant,
	authz *oauth.Authorization,
	session *session.IDPSession,
) (protocol.TokenResponse, error) {
	// TODO(oauth): create grants and issue tokens
	return nil, nil
}

func verifyPKCE(challenge, verifier string) bool {
	verifierHash := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(verifierHash[:])
	return subtle.ConstantTimeCompare([]byte(challenge), []byte(expectedChallenge)) == 1
}
