package upstreamapp

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	SessionInfo *session.Info
}

func WithSessionInfo(ctx context.Context, info *session.Info) context.Context {
	val := &contextValue{
		SessionInfo: info,
	}
	return context.WithValue(ctx, contextKey, val)
}

func getContextValue(ctx context.Context) *contextValue {
	val, _ := ctx.Value(contextKey).(*contextValue)
	return val
}

func GetValidSessionInfo(ctx context.Context) *session.Info {
	val := getContextValue(ctx)
	if val == nil || val.SessionInfo == nil || !val.SessionInfo.IsValid {
		return nil
	}
	return val.SessionInfo
}
