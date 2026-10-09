package authenticationflow

import (
	"context"
)

type contextKeyTypeOAuthSessionID struct{}

var contextKeyOAuthSessionID = contextKeyTypeOAuthSessionID{}

func GetOAuthSessionID(ctx context.Context) string {
	return ctx.Value(contextKeyOAuthSessionID).(string)
}

type contextKeyTypeSAMLSessionID struct{}

var contextKeySAMLSessionID = contextKeyTypeSAMLSessionID{}

func GetSAMLSessionID(ctx context.Context) string {
	return ctx.Value(contextKeySAMLSessionID).(string)
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

type contextKeyTypeIDTokenHintSID struct{}

var contextKeyIDTokenHintSID = contextKeyTypeIDTokenHintSID{}

func GetIDTokenHintSID(ctx context.Context) string {
	return ctx.Value(contextKeyIDTokenHintSID).(string)
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

type contextKeyTypeFormInputConsumption struct{}

var contextKeyFormInputConsumption = contextKeyTypeFormInputConsumption{}

type formInputConsumption struct{ consumed bool }

// WithFormInputConsumption installs a per-request marker.
// Session.MakeContext does it, and every entry of Service makes its context with it,
// so callers do not have to.
//
// Accept offers the same raw input to every input reactor it reaches,
// and the loop runs Accept again with the same input after ErrPauseAndRetryAccept.
// Without the marker, a raw input such as {"fields": []} would be taken by every consecutive form.
func WithFormInputConsumption(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyFormInputConsumption, &formInputConsumption{})
}

// MarkFormInputConsumed records that a form consumed the request's raw input.
func MarkFormInputConsumed(ctx context.Context) {
	getFormInputConsumption(ctx).consumed = true
}

// IsFormInputConsumed reports whether a form already consumed the request's raw input.
func IsFormInputConsumed(ctx context.Context) bool {
	return getFormInputConsumption(ctx).consumed
}

// getFormInputConsumption panics when the marker is missing, like the other getters here.
// Silently treating it as not consumed would disable the guard.
func getFormInputConsumption(ctx context.Context) *formInputConsumption {
	return ctx.Value(contextKeyFormInputConsumption).(*formInputConsumption)
}

type contextKeyTypeSession struct{}

var contextKeySession = contextKeyTypeSession{}

// GetSession returns the live *Session stored by MakeContext.
// Because a pointer is stored, mutations made to the session after MakeContext are visible.
func GetSession(ctx context.Context) *Session {
	return ctx.Value(contextKeySession).(*Session)
}
