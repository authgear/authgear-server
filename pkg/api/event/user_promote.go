package event

import "github.com/authgear/authgear-server/pkg/api/model"

const (
	BeforeUserPromote Type = "before_user_promote"
	AfterUserPromote  Type = "after_user_promote"
)

/*
	@Callback
		@Operation POST /before_user_create - Before user creation
			A user is about to be created.
			@RequestBody
				@JSONSchema {BeforeUserCreateEvent}
			@Response 200 {HookResponse}

		@Operation POST /after_user_create - After user creation
			A user is created.
			@RequestBody
				@JSONSchema {AfterUserCreateEvent}
			@Response 200 {EmptyResponse}
*/
type UserPromoteEvent struct {
	AnonymousUser model.User       `json:"anonymous_user"`
	User          model.User       `json:"user"`
	Identities    []model.Identity `json:"identities"`
}

// @JSONSchema
const BeforeUserPromoteEventSchema = `
{
	"$id": "#BeforeUserPromoteEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_user_promote"] },
		"payload": { "$ref": "#UserPromoteEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const AfterUserPromoteEventSchema = `
{
	"$id": "#AfterUserPromoteEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_user_promote"] },
		"payload": { "$ref": "#UserPromoteEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const UserPromoteEventPayloadSchema = `
{
	"$id": "#UserPromoteEventPayload",
	"type": "object",
	"properties": {
		"anonymous_user": { "$ref": "#User" },
		"user": { "$ref": "#User" },
		"identity": {
			"type": "array",
			"items": { "$ref": "#Identity" }
		}
	}
}
`

func (e *UserPromoteEvent) BeforeEventType() Type {
	return BeforeUserPromote
}

func (e *UserPromoteEvent) AfterEventType() Type {
	return AfterUserPromote
}

func (e *UserPromoteEvent) UserID() string {
	return e.User.ID
}
