package authenticationflow

import (
	"context"
)

type contextKeyTypeOAuthSessionID struct{}

var contextKeyOAuthSessionID = contextKeyTypeOAuthSessionID{}

func GetOAuthSessionID(ctx context.Context) string {
	return ctx.Value(contextKeyOAuthSessionID).(string)
}

type contextKeyTypeBotProtectionVerificationResult struct{}

var contextKeyBotProtectionVerificationResult = contextKeyTypeBotProtectionVerificationResult{}

func GetBotProtectionVerificationResult(ctx context.Context) *BotProtectionVerificationResult {

	result, ok := ctx.Value(contextKeyBotProtectionVerificationResult).(*BotProtectionVerificationResult)
	if !ok {
		return nil
	}
	return result
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
