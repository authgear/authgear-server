package event

import "github.com/authgear/authgear-server/pkg/api/model"

const (
	BeforeIdentityDelete Type = "before_identity_delete"
	AfterIdentityDelete  Type = "after_identity_delete"
)

/*
	@Callback
		@Operation POST /before_identity_delete - Before identity deletion
			An identity is about to be deleted.
			@RequestBody
				@JSONSchema {BeforeIdentityDeleteEvent}
			@Response 200 {HookResponse}

		@Operation POST /after_identity_delete - After identity deletion
			An identity is deleted.
			@RequestBody
				@JSONSchema {AfterIdentityDeleteEvent}
			@Response 200 {EmptyResponse}
*/
type IdentityDeleteEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

// @JSONSchema
const BeforeIdentityDeleteEventSchema = `
{
	"$id": "#BeforeIdentityDeleteEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_identity_delete"] },
		"payload": { "$ref": "#IdentityDeleteEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const AfterIdentityDeleteEventSchema = `
{
	"$id": "#AfterIdentityDeleteEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_identity_delete"] },
		"payload": { "$ref": "#IdentityDeleteEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const IdentityDeleteEventPayloadSchema = `
{
	"$id": "#IdentityDeleteEventPayload",
	"type": "object",
	"properties": {
		"user": { "$ref": "#User" },
		"identity": { "$ref": "#Identity" }
	}
}
`

func (e *IdentityDeleteEvent) BeforeEventType() Type {
	return BeforeIdentityDelete
}

func (e *IdentityDeleteEvent) AfterEventType() Type {
	return AfterIdentityDelete
}

func (e *IdentityDeleteEvent) UserID() string {
	return e.User.ID
}
