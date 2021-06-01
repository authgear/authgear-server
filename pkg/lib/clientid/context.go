package clientid

import (
	"context"
)

type clientIDContextKeyType struct{}

var clientIDContextKey = clientIDContextKeyType{}

type clientIDContext struct {
	ClientID string
}

func WithClientID(ctx context.Context, clientID string) context.Context {
	v, ok := ctx.Value(clientIDContextKey).(*clientIDContext)
	if ok {
		v.ClientID = clientID
		return ctx
	}

	return context.WithValue(ctx, clientIDContextKey, &clientIDContext{
		ClientID: clientID,
	})
}

func GetClientID(ctx context.Context) string {
	v, ok := ctx.Value(clientIDContextKey).(*clientIDContext)
	if ok {
		return v.ClientID
	}
	return ""
}
