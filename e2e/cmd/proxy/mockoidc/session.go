package mockoidc

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Session struct {
	SessionID string
	Scopes    []string
	User      User
}

type SessionStore struct {
	Store map[string]*Session
}

type IDTokenClaims struct {
	// UPN is specific to the Azure AD OIDC implementation
	// https://github.com/authgear/authgear-server/blob/2f147b2e1d314f26d5980e8e70c1f52501545c82/pkg/lib/authn/sso/adfs.go#L96
	UPN string `json:"upn,omitempty"`
	*jwt.RegisteredClaims
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		Store: make(map[string]*Session),
	}
}

func (ss *SessionStore) NewSession(clientID string, scope string, user User) (*Session, error) {
	// Use the ClientID as the session ID since e2e tests will not interact to get the code
	sessionID := clientID

	session := &Session{
		SessionID: sessionID,
		Scopes:    strings.Split(scope, " "),
		User:      user,
	}
	ss.Store[sessionID] = session

	return session, nil
}

func (ss *SessionStore) GetSessionByID(id string) (*Session, error) {
	session, ok := ss.Store[id]
	if !ok {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (ss *SessionStore) GetSessionByToken(token *jwt.Token) (*Session, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	sessionID := claims["jti"].(string)
	return ss.GetSessionByID(sessionID)
}

func (s *Session) AccessToken(config *Config, kp *Keypair, now time.Time) (string, error) {
	claims := s.registeredClaims(config.Issuer, config.ClientID, config.AccessTTL, now)
	return kp.SignJWT(claims)
}

func (s *Session) RefreshToken(config *Config, kp *Keypair, now time.Time) (string, error) {
	claims := s.registeredClaims(config.Issuer, config.ClientID, config.RefreshTTL, now)
	return kp.SignJWT(claims)
}

func (s *Session) IDToken(config *Config, kp *Keypair, now time.Time) (string, error) {
	base := &IDTokenClaims{
		RegisteredClaims: s.registeredClaims(config.Issuer, config.ClientID, config.AccessTTL, now),
		UPN:              s.User.ID(),
	}
	claims, err := s.User.Claims(s.Scopes, base)
	if err != nil {
		return "", err
	}

	return kp.SignJWT(claims)
}

func (s *Session) registeredClaims(issuer string, clientID string, ttl time.Duration, now time.Time) *jwt.RegisteredClaims {
	return &jwt.RegisteredClaims{
		Audience:  jwt.ClaimStrings{clientID},
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		ID:        s.SessionID,
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    issuer,
		NotBefore: jwt.NewNumericDate(now),
		Subject:   s.User.ID(),
	}
}
