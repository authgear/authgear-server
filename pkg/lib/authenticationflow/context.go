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

type contextKeyTypeIDToken struct{}

var contextKeyIDToken = contextKeyTypeIDToken{}

func GetIDToken(ctx context.Context) string {
	return ctx.Value(contextKeyIDToken).(string)
}

type contextKeyTypeSuppressIDPSessionCookie struct{}

var contextKeySuppressIDPSessionCookie = contextKeyTypeSuppressIDPSessionCookie{}

func GetSuppressIDPSessionCookie(ctx context.Context) bool {
	return ctx.Value(contextKeySuppressIDPSessionCookie).(bool)
}

type contextKeyTypeUserIDHint struct{}

var contextKeyUserIDHint = contextKeyTypeUserIDHint{}

func GetUserIDHint(ctx context.Context) string {
	return ctx.Value(contextKeyUserIDHint).(string)
}

type contextKeyTypeLoginHint struct{}

var contextKeyLoginHint = contextKeyTypeLoginHint{}

func GetLoginHint(ctx context.Context) string {
	return ctx.Value(contextKeyLoginHint).(string)
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
