package testing

import (
	"github.com/skygeario/skygear-server/pkg/core/auth"
	authinfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	authtoken "github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	model "github.com/skygeario/skygear-server/pkg/core/model"
)

type MockContext struct {
	accessKeyType model.KeyType
	authInfo      *authinfo.AuthInfo
	token         *authtoken.Token
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

func (m *MockContext) Token() *authtoken.Token {
	return m.token
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
	m.token = &authtoken.Token{
		AuthInfoID:  userID,
		PrincipalID: principalID,
	}
	return m
}

func (m *MockContext) MarkVerified() *MockContext {
	m.authInfo.Verified = true
	return m
}
