package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	OIDCJWTPreCreate event.Type = "oidc.jwt.pre_create"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("OIDCJWTPreCreateHookResponse", `
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
	event.RegisterResponseSchemaValidator(OIDCJWTPreCreate, s.PartValidator("OIDCJWTPreCreateHookResponse"))
}

type OIDCJWT struct {
	Payload map[string]interface{} `json:"payload"`
}

type OIDCJWTPreCreateBlockingEventPayload struct {
	UserRef    model.UserRef    `json:"-" resolve:"user"`
	UserModel  model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	JWT        OIDCJWT          `json:"jwt"`
}

func (e *OIDCJWTPreCreateBlockingEventPayload) BlockingEventType() event.Type {
	return OIDCJWTPreCreate
}

func (e *OIDCJWTPreCreateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *OIDCJWTPreCreateBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *OIDCJWTPreCreateBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *OIDCJWTPreCreateBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	mutationsEverApplied := false
	if response.Mutations.JWT.Payload != nil {
		e.JWT.Payload = response.Mutations.JWT.Payload
		mutationsEverApplied = true
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: mutationsEverApplied}
}

func (e *OIDCJWTPreCreateBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	return nil
}

var _ event.BlockingPayload = &OIDCJWTPreCreateBlockingEventPayload{}
