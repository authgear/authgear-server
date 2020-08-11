package authn

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	IsInvalid bool
	Session   Session
}

func WithInvalidAuthn(ctx context.Context) context.Context {
	actx := &contextValue{
		IsInvalid: true,
	}
	return context.WithValue(ctx, contextKey, actx)
}

func WithAuthn(ctx context.Context, s Session) context.Context {
	actx := &contextValue{
		Session: s,
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

func GetUserID(ctx context.Context) *string {
	actx := getContext(ctx)
	if actx == nil || actx.Session == nil {
		return nil
	}
	return &actx.Session.AuthnAttrs().UserID
}
