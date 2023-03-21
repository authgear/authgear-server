package workflow

import (
	"context"
)

type contextKeyTypeSuppressIDPSessionCookie struct{}

var contextKeySuppressIDPSessionCookie = contextKeyTypeSuppressIDPSessionCookie{}

func GetSuppressIDPSessionCookie(ctx context.Context) bool {
	return ctx.Value(contextKeySuppressIDPSessionCookie).(bool)
}

type contextKeyTypeWorkflowID struct{}

var contextKeyWorkflowID = contextKeyTypeWorkflowID{}

func GetWorkflowID(ctx context.Context) string {
	return ctx.Value(contextKeyWorkflowID).(string)
}
