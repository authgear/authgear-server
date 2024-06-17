package session

import (
	"context"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	IsInvalid bool
	Session   ListableSession
}

func WithInvalidSession(ctx context.Context) context.Context {
	actx := &contextValue{
		IsInvalid: true,
	}
	return context.WithValue(ctx, contextKey, actx)
}

func WithSession(ctx context.Context, s ListableSession) context.Context {
	actx := &contextValue{
		Session: s,
	}
	return context.WithValue(ctx, contextKey, actx)
}

func getContext(ctx context.Context) *contextValue {
	actx, _ := ctx.Value(contextKey).(*contextValue)
	return actx
}

func HasValidSession(ctx context.Context) bool {
	actx := getContext(ctx)
	if actx == nil {
		return true
	}
	return !actx.IsInvalid
}

func GetSession(ctx context.Context) ListableSession {
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
	userID := actx.Session.GetAuthenticationInfo().UserID
	return &userID
}
