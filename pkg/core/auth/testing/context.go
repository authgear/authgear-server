package testing

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type MockContext struct {
	accessKey model.AccessKey
	authInfo  *authinfo.AuthInfo
	session   *auth.Session
}

var _ auth.ContextGetter = &MockContext{}

func NewMockContext() *MockContext {
	return &MockContext{accessKey: model.AccessKey{Type: model.APIAccessKeyType}}
}

func (m *MockContext) AccessKey() model.AccessKey {
	return m.accessKey
}

func (m *MockContext) AuthInfo() *authinfo.AuthInfo {
	return m.authInfo
}

func (m *MockContext) Session() *auth.Session {
	return m.session
}

func (m *MockContext) UseNoAccessKey() *MockContext {
	m.accessKey.Type = model.NoAccessKeyType
	return m
}

func (m *MockContext) UseMasterKey() *MockContext {
	m.accessKey.Type = model.MasterAccessKeyType
	return m
}

func (m *MockContext) UseUser(userID string, principalID string) *MockContext {
	m.authInfo = &authinfo.AuthInfo{
		ID:         userID,
		VerifyInfo: map[string]bool{},
	}
	m.session = &auth.Session{
		ID:          fmt.Sprintf("%s-%s", userID, principalID),
		UserID:      userID,
		PrincipalID: principalID,
		AccessToken: fmt.Sprintf("access-token-%s-%s", userID, principalID),
	}
	return m
}

func (m *MockContext) MarkVerified() *MockContext {
	m.authInfo.Verified = true
	return m
}
