package event

import "github.com/authgear/authgear-server/pkg/api/model"

const (
	BeforeUserCreate Type = "before_user_create"
	AfterUserCreate  Type = "after_user_create"
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
type UserCreateEvent struct {
	User       model.User       `json:"user"`
	Identities []model.Identity `json:"identities"`
}

// @JSONSchema
const BeforeUserCreateEventSchema = `
{
	"$id": "#BeforeUserCreateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_user_create"] },
		"payload": { "$ref": "#UserCreateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const AfterUserCreateEventSchema = `
{
	"$id": "#AfterUserCreateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_user_create"] },
		"payload": { "$ref": "#UserCreateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const UserCreateEventPayloadSchema = `
{
	"$id": "#UserCreateEventPayload",
	"type": "object",
	"properties": {
		"user": { "$ref": "#User" },
		"identity": {
			"type": "array",
			"items": { "$ref": "#Identity" }
		}
	}
}
`

func (e *UserCreateEvent) BeforeEventType() Type {
	return BeforeUserCreate
}

func (e *UserCreateEvent) AfterEventType() Type {
	return AfterUserCreate
}

func (e *UserCreateEvent) UserID() string {
	return e.User.ID
}
