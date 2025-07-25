package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AuthenticationPreAuthenticated event.Type = "authentication.pre_authenticated"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("AuthenticationPreAuthenticatedHookResponse", `
{
	"allOf": [
		{ "$ref": "#/$defs/BaseHookResponseSchema" },
		{
			"if": {
				"properties": {
					"is_allowed": { "const": true }
				}
			},
			"then": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"is_allowed": {},
					"constraints": {},
					"rate_limits": {}
				}
			}
		}
	]
}`)

	s.Instantiate()
	event.RegisterResponseSchemaValidator(AuthenticationPreAuthenticated, s.PartValidator("AuthenticationPreAuthenticatedHookResponse"))
}

type AuthenticationPreAuthenticatedBlockingEventPayload struct {
	AuthenticationContext event.AuthenticationContext `json:"authentication_context"`

	Constraints *event.Constraints `json:"-"`
	RateLimits  event.RateLimits   `json:"-"`
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
	if response.RateLimits != nil {
		e.RateLimits = response.RateLimits
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationPreAuthenticatedBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationPreAuthenticatedBlockingEventPayload{}
