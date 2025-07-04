package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AuthenticationPostIdentified event.Type = "authentication.post_identified"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("AuthenticationPostIdentifiedHookResponse", `
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
					"bot_protection": {},
					"rate_limits": {}
				}
			}
		}
	]
}`)

	s.Instantiate()
	event.RegisterResponseSchemaValidator(AuthenticationPostIdentified, s.PartValidator("AuthenticationPostIdentifiedHookResponse"))
}

type AuthenticationPostIdentifiedBlockingEventPayload struct {
	Identification        model.Identification        `json:"identification"`
	AuthenticationContext event.AuthenticationContext `json:"authentication_context"`

	Constraints               *event.Constraints               `json:"-"`
	BotProtectionRequirements *event.BotProtectionRequirements `json:"-"`
	RateLimits                event.RateLimits                 `json:"-"`
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) BlockingEventType() event.Type {
	return AuthenticationPostIdentified
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) UserID() string {
	return e.AuthenticationContext.User.ID
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	if response.Constraints != nil {
		e.Constraints = response.Constraints
	}
	if response.BotProtection != nil {
		e.BotProtectionRequirements = response.BotProtection
	}
	if response.RateLimits != nil {
		e.RateLimits = response.RateLimits
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationPostIdentifiedBlockingEventPayload{}
