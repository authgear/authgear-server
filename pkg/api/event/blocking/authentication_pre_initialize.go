package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
)

const (
	AuthenticationPreInitialize event.Type = "authentication.pre_initialize"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("AuthenticationPreInitializeHookResponse", `
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
					"bot_protection": {}
				}
			}
		}
	]
}`)

	s.Instantiate()
	event.RegisterResponseSchemaValidator(AuthenticationPreInitialize, s.PartValidator("AuthenticationPreInitializeHookResponse"))
}

type AuthenticationPreInitializeBlockingEventPayload struct {
	AuthenticationContext event.AuthenticationContext `json:"authentication_context"`

	Constraints               *event.Constraints               `json:"-"`
	BotProtectionRequirements *event.BotProtectionRequirements `json:"-"`
}

func (e *AuthenticationPreInitializeBlockingEventPayload) BlockingEventType() event.Type {
	return AuthenticationPreInitialize
}

func (e *AuthenticationPreInitializeBlockingEventPayload) UserID() string {
	return ""
}

func (e *AuthenticationPreInitializeBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationPreInitializeBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *AuthenticationPreInitializeBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	if response.Constraints != nil {
		e.Constraints = response.Constraints
	}
	if response.BotProtection != nil {
		e.BotProtectionRequirements = response.BotProtection
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationPreInitializeBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationPreInitializeBlockingEventPayload{}
