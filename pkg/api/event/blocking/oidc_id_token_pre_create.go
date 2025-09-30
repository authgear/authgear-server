package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	//nolint:gosec // G101
	OIDCIDTokenPreCreate event.Type = "oidc.id_token.pre_create"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	s.Add("OIDCIDTokenPreCreateHookResponse", `
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
	event.RegisterResponseSchemaValidator(OIDCIDTokenPreCreate, s.PartValidator("OIDCIDTokenPreCreateHookResponse"))
}

type OIDCIDToken struct {
	Payload map[string]interface{} `json:"payload"`
}

type OIDCIDTokenPreCreateBlockingEventPayload struct {
	UserRef    model.UserRef    `json:"-" resolve:"user"`
	UserModel  model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
	JWT        OIDCIDToken      `json:"jwt"`
}

func (e *OIDCIDTokenPreCreateBlockingEventPayload) BlockingEventType() event.Type {
	return OIDCIDTokenPreCreate
}

func (e *OIDCIDTokenPreCreateBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *OIDCIDTokenPreCreateBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *OIDCIDTokenPreCreateBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *OIDCIDTokenPreCreateBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	mutationsEverApplied := false
	if response.Mutations.JWT.Payload != nil {
		e.JWT.Payload = response.Mutations.JWT.Payload
		mutationsEverApplied = true
	}
	return event.ApplyHookResponseResult{MutationsEverApplied: mutationsEverApplied}
}

func (e *OIDCIDTokenPreCreateBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	return nil
}

var _ event.BlockingPayload = &OIDCIDTokenPreCreateBlockingEventPayload{}
