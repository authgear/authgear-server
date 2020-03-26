package authn

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	IsInvalid bool
	Session   Session
	User      *UserInfo
}

func WithInvalidAuthn(ctx context.Context) context.Context {
	actx := &contextValue{
		IsInvalid: true,
	}
	return context.WithValue(ctx, contextKey, actx)
}

func WithAuthn(ctx context.Context, s Session, u *UserInfo) context.Context {
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

func GetUser(ctx context.Context) *UserInfo {
	actx := getContext(ctx)
	if actx == nil || actx.User == nil {
		return nil
	}
	return actx.User
}
