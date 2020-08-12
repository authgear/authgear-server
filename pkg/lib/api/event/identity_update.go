package event

import "github.com/authgear/authgear-server/pkg/auth/model"

const (
	BeforeIdentityUpdate Type = "before_identity_update"
	AfterIdentityUpdate  Type = "after_identity_update"
)

/*
	@Callback
		@Operation POST /before_identity_update - Before identity update
			An identity is about to be updated.
			@RequestBody
				@JSONSchema {BeforeIdentityUpdateEvent}
			@Response 200 {HookResponse}

		@Operation POST /after_identity_update - After identity update
			An identity is updated.
			@RequestBody
				@JSONSchema {AfterIdentityUpdateEvent}
			@Response 200 {EmptyResponse}
*/
type IdentityUpdateEvent struct {
	User        model.User     `json:"user"`
	NewIdentity model.Identity `json:"new_identity"`
	OldIdentity model.Identity `json:"old_identity"`
}

// @JSONSchema
const BeforeIdentityUpdateEventSchema = `
{
	"$id": "#BeforeIdentityUpdateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_identity_update"] },
		"payload": { "$ref": "#IdentityUpdateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const AfterIdentityUpdateEventSchema = `
{
	"$id": "#AfterIdentityUpdateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_identity_update"] },
		"payload": { "$ref": "#IdentityUpdateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const IdentityUpdateEventPayloadSchema = `
{
	"$id": "#IdentityUpdateEventPayload",
	"type": "object",
	"properties": {
		"user": { "$ref": "#User" },
		"old_identity": { "$ref": "#Identity" },
		"new_identity": { "$ref": "#Identity" }
	}
}
`

func (e *IdentityUpdateEvent) BeforeEventType() Type {
	return BeforeIdentityUpdate
}

func (e *IdentityUpdateEvent) AfterEventType() Type {
	return AfterIdentityUpdate
}

func (e *IdentityUpdateEvent) UserID() string {
	return e.User.ID
}
