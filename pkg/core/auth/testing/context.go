package testing

import (
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type MockContext struct {
	accessKeyType model.KeyType
	authInfo      *authinfo.AuthInfo
	session       *session.Session
}

var _ auth.ContextGetter = &MockContext{}

func NewMockContext() *MockContext {
	return &MockContext{accessKeyType: model.APIAccessKey}
}

func (m *MockContext) AccessKeyType() model.KeyType {
	return m.accessKeyType
}

func (m *MockContext) AuthInfo() *authinfo.AuthInfo {
	return m.authInfo
}

func (m *MockContext) Session() *session.Session {
	return m.session
}

func (m *MockContext) UseNoAccessKey() *MockContext {
	m.accessKeyType = model.NoAccessKey
	return m
}

func (m *MockContext) UseMasterKey() *MockContext {
	m.accessKeyType = model.MasterAccessKey
	return m
}

func (m *MockContext) UseUser(userID string, principalID string) *MockContext {
	m.authInfo = &authinfo.AuthInfo{
		ID:         userID,
		VerifyInfo: map[string]bool{},
	}
	m.session = &session.Session{
		UserID:      userID,
		PrincipalID: principalID,
	}
	return m
}

func (m *MockContext) MarkVerified() *MockContext {
	m.authInfo.Verified = true
	return m
}
