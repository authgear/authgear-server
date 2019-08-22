package session

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/time"
)

type MockProvider struct {
	Time    time.Provider
	counter int

	Sessions map[string]Session
}

var _ Provider = &MockProvider{}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		Time:     &time.MockProvider{},
		Sessions: map[string]Session{},
	}
}

func (p *MockProvider) Create(userID string, principalID string) (s *Session, err error) {
	now := p.Time.NowUTC()
	id := fmt.Sprintf("%s-%s-%d", userID, principalID, p.counter)
	sess := Session{
		ID:          id,
		UserID:      userID,
		PrincipalID: principalID,

		CreatedAt:            now,
		AccessedAt:           now,
		AccessToken:          "access-token-" + id,
		AccessTokenCreatedAt: now,
	}
	p.counter++

	p.Sessions[sess.ID] = sess

	return &sess, nil
}

func (p *MockProvider) GetByToken(token string, kind TokenKind) (*Session, error) {
	for _, s := range p.Sessions {
		var expectedToken string
		switch kind {
		case TokenKindAccessToken:
			expectedToken = s.AccessToken
		default:
			continue
		}

		if expectedToken != token {
			continue
		}

		return &s, nil
	}
	return nil, ErrSessionNotFound
}

func (p *MockProvider) Access(s *Session) error {
	s.AccessedAt = p.Time.NowUTC()
	p.Sessions[s.ID] = *s
	return nil
}

func (p *MockProvider) Invalidate(id string) error {
	delete(p.Sessions, id)
	return nil
}
