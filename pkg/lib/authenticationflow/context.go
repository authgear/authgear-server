package authenticationflow

import (
	"context"
)

type contextKeyTypeSuppressIDPSessionCookie struct{}

var contextKeySuppressIDPSessionCookie = contextKeyTypeSuppressIDPSessionCookie{}

func GetSuppressIDPSessionCookie(ctx context.Context) bool {
	return ctx.Value(contextKeySuppressIDPSessionCookie).(bool)
}

type contextKeyTypeFlowID struct{}

var contextKeyFlowID = contextKeyTypeFlowID{}

func GetFlowID(ctx context.Context) string {
	return ctx.Value(contextKeyFlowID).(string)
}
