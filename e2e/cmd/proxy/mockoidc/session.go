package mockoidc

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Session struct {
	SessionID string
	Scopes    []string
	User      User
}

type SessionStore struct {
	Store map[string]*Session
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

func (ss *SessionStore) GetSessionByToken(token jwt.Token) (*Session, error) {
	sessionIDIface, ok := token.Get(jwt.JwtIDKey)
	if !ok {
		return nil, errors.New("jti not found")
	}

	sessionID, ok := sessionIDIface.(string)
	if !ok {
		return nil, fmt.Errorf("jti is not a string: %T", sessionIDIface)
	}

	return ss.GetSessionByID(sessionID)
}

func (s *Session) AccessToken(config *Config, kp *Keypair, now time.Time) (string, error) {
	return s.signAnyToken(config.AccessTTL, config, kp, now)
}

func (s *Session) RefreshToken(config *Config, kp *Keypair, now time.Time) (string, error) {
	return s.signAnyToken(config.RefreshTTL, config, kp, now)
}

func (s *Session) IDToken(config *Config, kp *Keypair, now time.Time) (string, error) {
	return s.signAnyToken(config.AccessTTL, config, kp, now)
}

func (s *Session) signAnyToken(ttl time.Duration, config *Config, kp *Keypair, now time.Time) (string, error) {
	token := jwt.New()

	err := token.Set("upn", s.User.ID())
	if err != nil {
		return "", err
	}

	err = token.Set(jwt.AudienceKey, []string{config.ClientID})
	if err != nil {
		return "", err
	}

	err = token.Set(jwt.ExpirationKey, now.Add(ttl))
	if err != nil {
		return "", err
	}

	err = token.Set(jwt.JwtIDKey, s.SessionID)
	if err != nil {
		return "", err
	}

	err = token.Set(jwt.IssuedAtKey, now)
	if err != nil {
		return "", err
	}

	err = token.Set(jwt.IssuerKey, config.Issuer)
	if err != nil {
		return "", err
	}

	err = token.Set(jwt.NotBeforeKey, now)
	if err != nil {
		return "", err
	}

	err = token.Set(jwt.SubjectKey, s.User.ID())
	if err != nil {
		return "", err
	}

	err = s.User.AddClaims(s.Scopes, token)
	if err != nil {
		return "", err
	}

	return kp.SignJWT(token)
}
