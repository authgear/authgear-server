package auth

import (
	authinfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	authtoken "github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	role "github.com/skygeario/skygear-server/pkg/core/auth/role"
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
	container := &contextContainer{
		accessKeyType: model.APIAccessKey,
		authInfo: &authinfo.AuthInfo{
			ID:    "faseng.cat.id",
			Roles: []string{"user"},
		},
		roles: []role.Role{
			role.Role{
				Name: "user",
			},
		},
		token: &authtoken.Token{
			AccessToken: "faseng_access_token",
		},
	}
	return &MockContext{container: container}
}

// NewMockContextGetterWithAdminUser creates auth context with admin user
func NewMockContextGetterWithAdminUser() ContextGetter {
	container := &contextContainer{
		accessKeyType: model.APIAccessKey,
		authInfo: &authinfo.AuthInfo{
			ID:    "chima.cat.id",
			Roles: []string{"admin"},
		},
		roles: []role.Role{
			role.Role{
				Name:    "admin",
				IsAdmin: true,
			},
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

// AccessKeyType returns access key type from mock context
func (m *MockContext) AccessKeyType() model.KeyType {
	return m.container.accessKeyType
}

// AuthInfo returns auth info from mock context
func (m *MockContext) AuthInfo() *authinfo.AuthInfo {
	return m.container.authInfo
}

// Roles returns roles from mock context
func (m *MockContext) Roles() []role.Role {
	return m.container.roles
}

// Token returns token from mock context
func (m *MockContext) Token() *authtoken.Token {
	return m.container.token
}
