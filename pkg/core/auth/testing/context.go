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
	err       error
}

var (
	_ auth.ContextGetter = &MockContext{}
	_ auth.ContextSetter = &MockContext{}
)

func NewMockContext() *MockContext {
	return &MockContext{accessKey: model.AccessKey{Type: model.APIAccessKeyType}}
}

func (m *MockContext) AccessKey() model.AccessKey {
	return m.accessKey
}

func (m *MockContext) AuthInfo() (*authinfo.AuthInfo, error) {
	return m.authInfo, m.err
}

func (m *MockContext) MustAuthInfo() *authinfo.AuthInfo {
	return m.authInfo
}

func (m *MockContext) Session() (*auth.Session, error) {
	return m.session, m.err
}

func (m *MockContext) SetAccessKey(key model.AccessKey) {
	m.accessKey = key
}

func (m *MockContext) SetSessionAndAuthInfo(sess *auth.Session, info *authinfo.AuthInfo, err error) {
	m.session = sess
	m.authInfo = info
	m.err = err
}

func (m *MockContext) UseAPIAccessKey(clientID string) *MockContext {
	m.accessKey.Type = model.APIAccessKeyType
	m.accessKey.ClientID = clientID
	return m
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
		ID:              fmt.Sprintf("%s-%s", userID, principalID),
		UserID:          userID,
		PrincipalID:     principalID,
		AccessTokenHash: fmt.Sprintf("access-token-%s-%s", userID, principalID),
	}
	return m
}

func (m *MockContext) UseSession(sess *auth.Session) *MockContext {
	m.session = sess
	return m
}

func (m *MockContext) MarkVerified() *MockContext {
	m.authInfo.Verified = true
	return m
}

func (m *MockContext) SetVerifyInfo(info map[string]bool) *MockContext {
	m.authInfo.VerifyInfo = info
	return m
}

func (m *MockContext) CopyTo(setter auth.ContextSetter) {
	setter.SetAccessKey(m.accessKey)
	setter.SetSessionAndAuthInfo(m.session, m.authInfo, m.err)
}
