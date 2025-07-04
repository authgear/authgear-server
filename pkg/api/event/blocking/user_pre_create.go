package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	UserPreCreate event.Type = "user.pre_create"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("UserPreCreateHookResponse", `
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
	event.RegisterResponseSchemaValidator(UserPreCreate, s.PartValidator("UserPreCreateHookResponse"))
}

type UserPreCreateBlockingEventPayload struct {
	UserRef    model.UserRef    `json:"-" resolve:"user"`
	UserModel  model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	AdminAPI   bool             `json:"-"`
}

func (e *UserPreCreateBlockingEventPayload) BlockingEventType() event.Type {
	return UserPreCreate
}

func (e *UserPreCreateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *UserPreCreateBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	if e.AdminAPI {
		return event.TriggeredByTypeAdminAPI
	}
	return event.TriggeredByTypeUser
}

func (e *UserPreCreateBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *UserPreCreateBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	user, mutated := ApplyUserMutations(e.UserModel, response.Mutations.User)
	if mutated {
		e.UserModel = user
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: mutated}
}

func (e *UserPreCreateBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	userID := e.UserID()
	userMutations := MakeUserMutations(e.UserModel)
	return PerformEffectsOnUser(ctx, effectCtx, userID, userMutations)
}

var _ event.BlockingPayload = &UserPreCreateBlockingEventPayload{}
