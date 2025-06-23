package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreScheduleAnonymization event.Type = "user.pre_schedule_anonymization"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("UserPreScheduleAnonymizationHookResponse", `
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
	event.RegisterResponseSchemaValidator(UserPreScheduleAnonymization, s.PartValidator("UserPreScheduleAnonymizationHookResponse"))
}

type UserPreScheduleAnonymizationBlockingEventPayload struct {
	UserRef   model.UserRef `json:"-" resolve:"user"`
	UserModel model.User    `json:"user"`
	AdminAPI  bool          `json:"-"`
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) BlockingEventType() event.Type {
	return UserPreScheduleAnonymization
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	user, mutated := ApplyUserMutations(e.UserModel, response.Mutations.User)
	if mutated {
		e.UserModel = user
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: mutated}
}

func (e *UserPreScheduleAnonymizationBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	userID := e.UserID()
	userMutations := MakeUserMutations(e.UserModel)
	return PerformEffectsOnUser(ctx, effectCtx, userID, userMutations)
}

var _ event.BlockingPayload = &UserPreScheduleAnonymizationBlockingEventPayload{}
