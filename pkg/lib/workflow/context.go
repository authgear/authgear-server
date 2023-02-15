package workflow

import (
	"context"
)

type contextKeyTypeClientID struct{}

var contextKeyClientID = contextKeyTypeClientID{}

func GetClientID(ctx context.Context) string {
	return ctx.Value(contextKeyClientID).(string)
}

type contextKeyTypeSuppressIDPSessionCookie struct{}

var contextKeySuppressIDPSessionCookie = contextKeyTypeSuppressIDPSessionCookie{}

func GetSuppressIDPSessionCookie(ctx context.Context) bool {
	return ctx.Value(contextKeySuppressIDPSessionCookie).(bool)
}

type contextKeyTypeState struct{}

var contextKeyState = contextKeyTypeState{}

func GetState(ctx context.Context) string {
	return ctx.Value(contextKeyState).(string)
}

type contextKeyTypeWorkflowID struct{}

var contextKeyWorkflowID = contextKeyTypeWorkflowID{}

func GetWorkflowID(ctx context.Context) string {
	return ctx.Value(contextKeyWorkflowID).(string)
}
