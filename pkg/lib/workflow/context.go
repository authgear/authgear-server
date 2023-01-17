package workflow

import (
	"context"
)

type contextKeyTypeClientID struct{}

var contextKeyClientID = contextKeyTypeClientID{}

func GetClientID(ctx context.Context) string {
	return ctx.Value(contextKeyClientID).(string)
}
