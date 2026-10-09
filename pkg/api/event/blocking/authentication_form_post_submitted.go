package blocking

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
)

const (
	AuthenticationFormPostSubmitted event.Type = "authentication.form.post_submitted"
)

func init() {
	s := event.GetBaseHookResponseSchema()
	// It does not reference BaseHookResponseSchema,
	// because the is_allowed: false branch there has additionalProperties: false, which rejects reasons.
	s.Add("FormHookResponseReason", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["type", "field", "message"],
	"properties": {
		"type": { "type": "string", "enum": ["invalid_form_field"] },
		"field": {
			"type": "object",
			"additionalProperties": false,
			"required": ["user_profile"],
			"properties": {
				"user_profile": {
					"type": "object",
					"additionalProperties": false,
					"required": ["pointer"],
					"properties": {
						"pointer": { "type": "string" }
					}
				}
			}
		},
		"message": { "type": "string" }
	}
}`)
	s.Add("AuthenticationFormPostSubmittedHookResponse", `
{
	"oneOf": [
		{
			"type": "object",
			"additionalProperties": false,
			"required": ["is_allowed"],
			"properties": {
				"is_allowed": { "const": true }
			}
		},
		{
			"type": "object",
			"additionalProperties": false,
			"required": ["is_allowed"],
			"properties": {
				"is_allowed": { "const": false },
				"title": { "type": "string" },
				"reason": { "type": "string" },
				"reasons": {
					"type": "array",
					"items": { "$ref": "#/$defs/FormHookResponseReason" }
				}
			}
		}
	]
}`)

	s.Instantiate()
	event.RegisterResponseSchemaValidator(AuthenticationFormPostSubmitted, s.PartValidator("AuthenticationFormPostSubmittedHookResponse"))
}

type AuthenticationFormPostSubmittedBlockingEventPayload struct {
	UserRef               model.UserRef               `json:"-" resolve:"user"`
	UserModel             model.User                  `json:"user"`
	AuthenticationContext event.AuthenticationContext `json:"authentication_context"`
	Form                  event.SubmittedForm         `json:"form"`
}

func (e *AuthenticationFormPostSubmittedBlockingEventPayload) BlockingEventType() event.Type {
	return AuthenticationFormPostSubmitted
}

func (e *AuthenticationFormPostSubmittedBlockingEventPayload) UserID() string {
	return e.UserRef.ID
}

func (e *AuthenticationFormPostSubmittedBlockingEventPayload) GetTriggeredBy() event.TriggeredByType {
	return event.TriggeredByTypeUser
}

func (e *AuthenticationFormPostSubmittedBlockingEventPayload) FillContext(ctx *event.Context) {}

func (e *AuthenticationFormPostSubmittedBlockingEventPayload) ApplyHookResponse(ctx context.Context, response event.HookResponse) event.ApplyHookResponseResult {
	return event.ApplyHookResponseResult{MutationsEverApplied: false}
}

func (e *AuthenticationFormPostSubmittedBlockingEventPayload) PerformEffects(ctx context.Context, effectCtx event.MutationsEffectContext) error {
	// no-op
	return nil
}

var _ event.BlockingPayload = &AuthenticationFormPostSubmittedBlockingEventPayload{}
