package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AuthenticationPostIdentified event.Type = "authentication.post_identified"
)

type AuthenticationPostIdentifiedBlockingEventPayload struct {
	Identification model.Identification        `json:"identification"`
	Authentication event.AuthenticationContext `json:"authentication"`

	Constraints               *event.Constraints               `json:"-"`
	BotProtectionRequirements *event.BotProtectionRequirements `json:"-"`
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) BlockingEventType() event.Type {
	return AuthenticationPostIdentified
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) UserID() string {
	return e.Authentication.User.ID
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
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationPostIdentifiedBlockingEventPayload{}
