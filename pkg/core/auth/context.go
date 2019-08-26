package auth

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/model"
)

type contextKey string

var (
	keyContainer = contextKey("container")
)

// ContextGetter provides interface for getting authentication data
type ContextGetter interface {
	AccessKey() model.AccessKey
	AuthInfo() *authinfo.AuthInfo
	Session() *session.Session
}

// ContextSetter provides interface for setting authentication data
type ContextSetter interface {
	SetAccessKey(model.AccessKey)
	SetAuthInfo(*authinfo.AuthInfo)
	SetSession(*session.Session)
}

// TODO: handle thread safety
type contextContainer struct {
	accessKey model.AccessKey
	authInfo  *authinfo.AuthInfo
	session   *session.Session
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

func (a *authContext) AccessKey() model.AccessKey {
	container := a.container()
	return container.accessKey
}

func (a *authContext) AuthInfo() *authinfo.AuthInfo {
	container := a.container()
	return container.authInfo
}

func (a *authContext) Session() *session.Session {
	container := a.container()
	return container.session
}

func (a *authContext) SetAccessKey(key model.AccessKey) {
	container := a.container()
	container.accessKey = key
}

func (a *authContext) SetAuthInfo(authInfo *authinfo.AuthInfo) {
	container := a.container()
	container.authInfo = authInfo
}

func (a *authContext) SetSession(session *session.Session) {
	container := a.container()
	container.session = session
}

func (a *authContext) container() *contextContainer {
	return a.Value(keyContainer).(*contextContainer)
}
