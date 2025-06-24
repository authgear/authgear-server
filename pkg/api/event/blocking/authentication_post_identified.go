package blocking

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
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
					"rate_limit": {}
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

	RateLimitReservations []*ratelimit.Reservation `json:"-"`

	Constraints               *event.Constraints               `json:"-"`
	BotProtectionRequirements *event.BotProtectionRequirements `json:"-"`
	RateLimit                 *event.RateLimit                 `json:"-"`
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
	result := event.ApplyHookResponseResult{}
	if response.Constraints != nil {
		e.Constraints = response.Constraints
	}
	if response.BotProtection != nil {
		e.BotProtectionRequirements = response.BotProtection
	}
	if response.RateLimit != nil {
		e.RateLimit = response.RateLimit
		result.RateLimitEverApplied = true
	}
	return result
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) PerformMutationEffects(ctx context.Context, eventCtx event.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) PerformRateLimitEffects(ctx context.Context, eventCtx event.Context, effectCtx event.RateLimitContext) error {
	if e.RateLimit == nil {
		return nil
	}
	var errs []error
	for _, resv := range e.RateLimitReservations {
		resv := resv
		_, failedReservation, err := effectCtx.RateLimiter.AdjustWeight(ctx, resv, e.RateLimit.Weight)
		if err != nil {
			errs = append(errs, err)
		} else if err := failedReservation.Error(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

var _ event.BlockingPayload = &AuthenticationPostIdentifiedBlockingEventPayload{}
