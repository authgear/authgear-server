package auth

import (
	authinfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	authtoken "github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	model "github.com/skygeario/skygear-server/pkg/core/model"
)

// MockContext is mock of authContext
type MockContext struct {
	container *contextContainer
}

// NewMockContextGetter create empty auth context
func NewMockContextGetter() ContextGetter {
	container := &contextContainer{}
	return &MockContext{container: container}
}

// NewMockContextGetterWithDefaultUser creates auth context with default user
func NewMockContextGetterWithDefaultUser() ContextGetter {
	return NewMockContextGetterWithUser("faseng.cat.id", true, map[string]bool{})
}

// NewMockContextGetterWithUnverifiedUser creates auth context with unverified user
func NewMockContextGetterWithUnverifiedUser(verifyInfo map[string]bool) ContextGetter {
	return NewMockContextGetterWithUser("faseng.cat.id", false, verifyInfo)
}

// NewMockContextGetterWithUser creates auth context with user
func NewMockContextGetterWithUser(userID string, verified bool, verifyInfo map[string]bool) ContextGetter {
	container := &contextContainer{
		accessKeyType: model.APIAccessKey,
		authInfo: &authinfo.AuthInfo{
			ID:         userID,
			Verified:   verified,
			VerifyInfo: verifyInfo,
		},
		token: &authtoken.Token{
			AccessToken: "faseng_access_token",
		},
	}
	return &MockContext{container: container}
}

// NewMockContextGetterWithMasterkeyDefaultUser creates auth context with default user and master key
func NewMockContextGetterWithMasterkeyDefaultUser() ContextGetter {
	ctx := NewMockContextGetterWithDefaultUser().(*MockContext)
	ctx.container.accessKeyType = model.MasterAccessKey
	return ctx
}

// NewMockContextGetterWithAdminUser creates auth context with admin user
func NewMockContextGetterWithAdminUser() ContextGetter {
	container := &contextContainer{
		accessKeyType: model.APIAccessKey,
		authInfo: &authinfo.AuthInfo{
			ID: "chima.cat.id",
		},
		token: &authtoken.Token{
			AccessToken: "chima_access_token",
		},
	}
	return &MockContext{container: container}
}

// NewMockContextGetterWithMasterKey creates auth context with master key
func NewMockContextGetterWithMasterKey() ContextGetter {
	container := &contextContainer{
		accessKeyType: model.MasterAccessKey,
	}
	return &MockContext{container: container}
}

// NewMockContextGetterWithAPIKey creates auth context with api key
func NewMockContextGetterWithAPIKey() ContextGetter {
	container := &contextContainer{
		accessKeyType: model.APIAccessKey,
	}
	return &MockContext{container: container}
}

// AccessKeyType returns access key type from mock context
func (m *MockContext) AccessKeyType() model.KeyType {
	return m.container.accessKeyType
}

// AuthInfo returns auth info from mock context
func (m *MockContext) AuthInfo() *authinfo.AuthInfo {
	return m.container.authInfo
}

// Token returns token from mock context
func (m *MockContext) Token() *authtoken.Token {
	return m.container.token
}
