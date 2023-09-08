package authenticationflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type contextKeyTypeOAuthSessionID struct{}

var contextKeyOAuthSessionID = contextKeyTypeOAuthSessionID{}

func GetOAuthSessionID(ctx context.Context) string {
	return ctx.Value(contextKeyOAuthSessionID).(string)
}

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

type contextKeyTypeFlowReference struct{}

var contextKeyFlowReference = contextKeyTypeFlowReference{}

func GetFlowReference(ctx context.Context) FlowReference {
	return ctx.Value(contextKeyFlowReference).(FlowReference)
}

type contextKeyTypeFlowRootObject struct{}

var contextKeyFlowRootObject = contextKeyTypeFlowRootObject{}

func GetFlowRootObject(ctx context.Context) config.AuthenticationFlowObject {
	v := ctx.Value(contextKeyFlowRootObject)
	if v == nil {
		return nil
	}
	return v.(config.AuthenticationFlowObject)
}
