package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AuthenticationPreAuthenticated event.Type = "authentication.pre_authenticated"
)

type AuthenticationPreAuthenticatedBlockingEventPayload struct {
	AuthenticationContext event.AuthenticationContext `json:"authentication_context"`

	Constraints *event.Constraints `json:"-"`
}

func (e *AuthenticationPreAuthenticatedBlockingEventPayload) BlockingEventType() event.Type {
	return AuthenticationPreAuthenticated
}

func (e *AuthenticationPreAuthenticatedBlockingEventPayload) UserID() string {
	return e.AuthenticationContext.User.ID
}

func (e *AuthenticationPreAuthenticatedBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationPreAuthenticatedBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *AuthenticationPreAuthenticatedBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	if response.Constraints != nil {
		e.Constraints = response.Constraints
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationPreAuthenticatedBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationPreAuthenticatedBlockingEventPayload{}
