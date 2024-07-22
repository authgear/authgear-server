package test

import (
	"context"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
)

type MockSession struct {
	Type  session.Type
	ID    string
	Attrs *session.Attrs
}

func NewMockSession() *MockSession {
	return &MockSession{
		Type: session.TypeIdentityProvider,
		ID:   "session-id",
		Attrs: &session.Attrs{
			UserID: "user-id",
		},
	}
}

func (m MockSession) Session() {}

func (m MockSession) SessionID() string { return m.ID }

func (m MockSession) SessionType() session.Type { return m.Type }

func (m MockSession) GetClientID() string { return "" }

func (m MockSession) GetCreatedAt() time.Time { return time.Time{} }

func (m MockSession) GetAuthenticatedAt() time.Time { return time.Time{} }

func (m MockSession) GetAccessInfo() *access.Info { return nil }

func (m MockSession) GetDeviceInfo() (map[string]interface{}, bool) { return nil, false }

func (m *MockSession) GetUserID() string { return m.Attrs.UserID }

func (m *MockSession) GetOIDCAMR() ([]string, bool) { return m.Attrs.GetAMR() }

func (m MockSession) ToAPIModel() *model.Session { return nil }

func (m MockSession) GetAuthenticationInfo() authenticationinfo.T {
	amr, _ := m.GetOIDCAMR()
	return authenticationinfo.T{
		UserID:          m.GetUserID(),
		AMR:             amr,
		AuthenticatedAt: m.GetAuthenticatedAt(),
	}
}

func (m *MockSession) GetAuthenticationInfoByThisSession() authenticationinfo.T {
	amr, _ := m.GetOIDCAMR()
	return authenticationinfo.T{
		UserID:                     m.GetUserID(),
		AMR:                        amr,
		AuthenticatedAt:            m.GetAuthenticatedAt(),
		AuthenticatedBySessionType: string(m.SessionType()),
		AuthenticatedBySessionID:   m.SessionID(),
	}
}

func (m *MockSession) SetUserID(id string) *MockSession {
	m.Attrs.UserID = id
	return m
}

func (m *MockSession) SetSessionID(id string) *MockSession {
	m.ID = id
	return m
}

func (m *MockSession) ToRequest(r *http.Request) *http.Request {
	return r.WithContext(m.ToContext(r.Context()))
}

func (m *MockSession) ToContext(ctx context.Context) context.Context {
	return session.WithSession(ctx, m)
}

func (s *MockSession) SSOGroupIDPSessionID() string {
	return ""
}

func (s *MockSession) IsSameSSOGroup(ss session.ListableSession) bool {
	return false
}

func (s *MockSession) Equal(ss session.ListableSession) bool {
	return false
}
