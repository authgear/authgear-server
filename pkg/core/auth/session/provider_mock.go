package session

import (
	"fmt"
	"sort"

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
		case auth.SessionTokenKindRefreshToken:
			expectedToken = s.RefreshToken
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

func (p *MockProvider) Get(id string) (*auth.Session, error) {
	session, ok := p.Sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

func (p *MockProvider) Access(s *auth.Session) error {
	s.AccessedAt = p.Time.NowUTC()
	p.Sessions[s.ID] = *s
	return nil
}

func (p *MockProvider) Invalidate(session *auth.Session) error {
	delete(p.Sessions, session.ID)
	return nil
}

func (p *MockProvider) InvalidateBatch(sessions []*auth.Session) error {
	for _, session := range sessions {
		delete(p.Sessions, session.ID)
	}
	return nil
}

func (p *MockProvider) InvalidateAll(userID string, sessionID string) error {
	for _, session := range p.Sessions {
		if session.UserID == userID && session.ID != sessionID {
			delete(p.Sessions, session.ID)
		}
	}
	return nil
}

func (p *MockProvider) List(userID string) (sessions []*auth.Session, err error) {
	for _, session := range p.Sessions {
		if session.UserID == userID {
			s := session
			sessions = append(sessions, &s)
		}
	}
	sort.Sort(sessionSlice(sessions))
	return
}

func (p *MockProvider) Refresh(session *auth.Session) error {
	session.AccessToken = fmt.Sprintf("access-token-%s-%d", session.ID, p.counter)
	p.Sessions[session.ID] = *session
	p.counter++
	return nil
}
