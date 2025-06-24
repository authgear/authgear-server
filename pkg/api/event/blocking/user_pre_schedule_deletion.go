package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreScheduleDeletion event.Type = "user.pre_schedule_deletion"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("UserPreScheduleDeletionHookResponse", `
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
					"mutations": {}
				}
			}
		}
	]
}`)

	s.Instantiate()
	event.RegisterResponseSchemaValidator(UserPreScheduleDeletion, s.PartValidator("UserPreScheduleDeletionHookResponse"))
}

type UserPreScheduleDeletionBlockingEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserPreScheduleDeletionBlockingEventPayload) BlockingEventType() event.Type {
	return UserPreScheduleDeletion
}

func (e *UserPreScheduleDeletionBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserPreScheduleDeletionBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserPreScheduleDeletionBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *UserPreScheduleDeletionBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	user, mutated := ApplyUserMutations(e.UserModel, response.Mutations.User)
	if mutated {
		e.UserModel = user
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: mutated}
}

func (e *UserPreScheduleDeletionBlockingEventPayload) PerformMutationEffects(ctx context.Context, eventCtx event.Context, effectCtx event.MutationsEffectContext) error {
	userID := e.UserID()
	userMutations := MakeUserMutations(e.UserModel)
	return PerformEffectsOnUser(ctx, effectCtx, userID, userMutations)
}

func (e *UserPreScheduleDeletionBlockingEventPayload) PerformRateLimitEffects(ctx context.Context, eventCtx event.Context, effectCtx event.RateLimitContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &UserPreScheduleDeletionBlockingEventPayload{}
