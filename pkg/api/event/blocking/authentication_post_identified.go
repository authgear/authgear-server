package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	AuthenticationPostIdentified event.Type = "authentication.post_identified"
)

type AuthenticationPostIdentifiedBlockingEventPayload struct {
	Identity       *model.Identity                         `json:"identity"`
	IDToken        *string                                 `json:"id_token"`
	Identification config.AuthenticationFlowIdentification `json:"identification"`
	Authentication event.AuthenticationContext             `json:"authentication"`

	Constraints *event.Constraints `json:"-"`
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) BlockingEventType() event.Type {
	return AuthenticationPostIdentified
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) UserID() string {
	return e.Identity.UserID
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	if response.Constraints != nil {
		e.Constraints = response.Constraints
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationPostIdentifiedBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationPostIdentifiedBlockingEventPayload{}
