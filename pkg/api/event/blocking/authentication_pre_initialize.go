package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AuthenticationPreInitialize event.Type = "authentication.pre_initialize"
)

type AuthenticationPreInitializeBlockingEventPayload struct {
	Authentication event.AuthenticationContext `json:"authentication"`

	Constraints *event.Constraints `json:"-"`
}

func (e *AuthenticationPreInitializeBlockingEventPayload) BlockingEventType() event.Type {
	return AuthenticationPreInitialize
}

func (e *AuthenticationPreInitializeBlockingEventPayload) UserID() string {
	return e.Authentication.User.ID
}

func (e *AuthenticationPreInitializeBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationPreInitializeBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *AuthenticationPreInitializeBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	if response.Constraints != nil {
		e.Constraints = response.Constraints
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationPreInitializeBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationPreInitializeBlockingEventPayload{}
