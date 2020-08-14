package event

import "github.com/authgear/authgear-server/pkg/lib/api/model"

const (
	BeforeIdentityCreate Type = "before_identity_create"
	AfterIdentityCreate  Type = "after_identity_create"
)

/*
	@Callback
		@Operation POST /before_identity_create - Before identity creation
			An identity is about to be created.
			@RequestBody
				@JSONSchema {BeforeIdentityCreateEvent}
			@Response 200 {HookResponse}

		@Operation POST /after_identity_create - After identity creation
			An identity is created.
			@RequestBody
				@JSONSchema {AfterIdentityCreateEvent}
			@Response 200 {EmptyResponse}
*/
type IdentityCreateEvent struct {
	User     model.User     `json:"user"`
	Identity model.Identity `json:"identity"`
}

// @JSONSchema
const BeforeIdentityCreateEventSchema = `
{
	"$id": "#BeforeIdentityCreateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_identity_create"] },
		"payload": { "$ref": "#IdentityCreateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const AfterIdentityCreateEventSchema = `
{
	"$id": "#AfterIdentityCreateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_identity_create"] },
		"payload": { "$ref": "#IdentityCreateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const IdentityCreateEventPayloadSchema = `
{
	"$id": "#IdentityCreateEventPayload",
	"type": "object",
	"properties": {
		"user": { "$ref": "#User" },
		"identity": { "$ref": "#Identity" }
	}
}
`

func (e *IdentityCreateEvent) BeforeEventType() Type {
	return BeforeIdentityCreate
}

func (e *IdentityCreateEvent) AfterEventType() Type {
	return AfterIdentityCreate
}

func (e *IdentityCreateEvent) UserID() string {
	return e.User.ID
}
