package authn

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	IsInvalid bool
	Session   Session
	User      *authinfo.AuthInfo
}

func WithInvalidAuthn(ctx context.Context) context.Context {
	actx := &contextValue{
		IsInvalid: true,
	}
	return context.WithValue(ctx, contextKey, actx)
}

func WithAuthn(ctx context.Context, s Session, u *authinfo.AuthInfo) context.Context {
	actx := &contextValue{
		Session: s,
		User:    u,
	}
	return context.WithValue(ctx, contextKey, actx)
}

func getContext(ctx context.Context) *contextValue {
	actx, _ := ctx.Value(contextKey).(*contextValue)
	return actx
}

func IsValidAuthn(ctx context.Context) bool {
	actx := getContext(ctx)
	if actx == nil {
		return true
	}
	return !actx.IsInvalid
}

func GetSession(ctx context.Context) Session {
	actx := getContext(ctx)
	if actx == nil || actx.Session == nil {
		return nil
	}
	return actx.Session
}

func GetAuthInfo(ctx context.Context) *authinfo.AuthInfo {
	actx := getContext(ctx)
	if actx == nil || actx.User == nil {
		return nil
	}
	return actx.User
}
