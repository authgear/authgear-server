package auth

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/auth/role"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type contextKey string

var (
	keyContainer = contextKey("container")
)

// ContextGetter provides interface for getting authentication data
type ContextGetter interface {
	AccessKeyType() model.KeyType
	AuthInfo() *authinfo.AuthInfo
	Roles() []role.Role
	Token() *authtoken.Token
}

// ContextSetter provides interface for setting authentication data
type ContextSetter interface {
	SetAccessKeyType(model.KeyType)
	SetAuthInfo(*authinfo.AuthInfo)
	SetRoles([]role.Role)
	SetToken(*authtoken.Token)
}

// TODO: handle thread safety
type contextContainer struct {
	accessKeyType model.KeyType
	authInfo      *authinfo.AuthInfo
	roles         []role.Role
	token         *authtoken.Token
}

type authContext struct {
	context.Context
}

func InitRequestAuthContext(req *http.Request) *http.Request {
	container := &contextContainer{}
	return req.WithContext(context.WithValue(req.Context(), keyContainer, container))
}

// NewContextGetterWithContext creates a new context.AuthGetter from context
func NewContextGetterWithContext(ctx context.Context) ContextGetter {
	return &authContext{Context: ctx}
}

// NewContextSetterWithContext creates a new context.AuthSetter from context
func NewContextSetterWithContext(ctx context.Context) ContextSetter {
	return &authContext{Context: ctx}
}

func (a *authContext) AccessKeyType() model.KeyType {
	container := a.container()
	return container.accessKeyType
}

func (a *authContext) AuthInfo() *authinfo.AuthInfo {
	container := a.container()
	return container.authInfo
}

func (a *authContext) Roles() []role.Role {
	container := a.container()
	return container.roles
}

func (a *authContext) Token() *authtoken.Token {
	container := a.container()
	return container.token
}

func (a *authContext) SetAccessKeyType(keyType model.KeyType) {
	container := a.container()
	container.accessKeyType = keyType
}

func (a *authContext) SetAuthInfo(authInfo *authinfo.AuthInfo) {
	container := a.container()
	container.authInfo = authInfo
}

func (a *authContext) SetRoles(roles []role.Role) {
	container := a.container()
	container.roles = roles
}

func (a *authContext) SetToken(token *authtoken.Token) {
	container := a.container()
	container.token = token
}

func (a *authContext) container() *contextContainer {
	return a.Value(keyContainer).(*contextContainer)
}
