package event

import "github.com/authgear/authgear-server/pkg/auth/model"

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

func (IdentityDeleteEvent) BeforeEventType() Type {
	return BeforeIdentityDelete
}

func (IdentityDeleteEvent) AfterEventType() Type {
	return AfterIdentityDelete
}

func (event IdentityDeleteEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	user := event.User
	mutations.ApplyToUser(&user)
	return IdentityDeleteEvent{
		User:     user,
		Identity: event.Identity,
	}
}

func (event IdentityDeleteEvent) UserID() string {
	return event.User.ID
}
