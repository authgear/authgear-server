package session

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type MockProvider struct {
	ClientID string
	Time     time.Provider
	counter  int

	Sessions map[string]auth.Session
}

var _ Provider = &MockProvider{}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		ClientID: "client-id",
		Time:     &time.MockProvider{},
		Sessions: map[string]auth.Session{},
	}
}

func (p *MockProvider) Create(userID string, principalID string) (s *auth.Session, err error) {
	now := p.Time.NowUTC()
	id := fmt.Sprintf("%s-%s-%d", userID, principalID, p.counter)
	sess := auth.Session{
		ID:          id,
		ClientID:    p.ClientID,
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

func (p *MockProvider) GetByToken(token string, kind auth.SessionTokenKind) (*auth.Session, error) {
	for _, s := range p.Sessions {
		var expectedToken string
		switch kind {
		case auth.SessionTokenKindAccessToken:
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

func (p *MockProvider) Access(s *auth.Session) error {
	s.AccessedAt = p.Time.NowUTC()
	p.Sessions[s.ID] = *s
	return nil
}

func (p *MockProvider) Invalidate(id string) error {
	delete(p.Sessions, id)
	return nil
}

func (p *MockProvider) Refresh(session *auth.Session) error {
	session.AccessToken = fmt.Sprintf("access-token-%s-%d", session.ID, p.counter)
	p.counter++
	return nil
}
