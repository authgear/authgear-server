package session

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	SessionInfo *model.SessionInfo
}

func WithSessionInfo(ctx context.Context, info *model.SessionInfo) context.Context {
	val := &contextValue{
		SessionInfo: info,
	}
	return context.WithValue(ctx, contextKey, val)
}

func getContextValue(ctx context.Context) *contextValue {
	val, _ := ctx.Value(contextKey).(*contextValue)
	return val
}

func GetValidSessionInfo(ctx context.Context) *model.SessionInfo {
	val := getContextValue(ctx)
	if val == nil || val.SessionInfo == nil || !val.SessionInfo.IsValid {
		return nil
	}
	return val.SessionInfo
}
