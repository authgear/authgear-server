package event

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

const (
	BeforeUserUpdate Type = "before_user_update"
	AfterUserUpdate  Type = "after_user_update"
)

type UserUpdateReason string

const (
	UserUpdateReasonUpdateMetadata = "update_metadata"
	UserUpdateReasonUpdateIdentity = "update_identity"
	UserUpdateReasonVerification   = "verification"
	UserUpdateReasonAdministrative = "administrative"
)

/*
	@Callback UserUpdateEvent
		@Operation POST /before_user_update - Before user update
			A user is about to be updated.
			@RequestBody
				@JSONSchema {BeforeUserUpdateEvent}
			@Response 200 {HookResponse}

		@Operation POST /after_user_update - After user update
			A user is updated.
			@RequestBody
				@JSONSchema {AfterUserUpdateEvent}
			@Response 200 {EmptyResponse}
*/
type UserUpdateEvent struct {
	Reason   UserUpdateReason        `json:"reason"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
	User     model.User              `json:"user"`
}

// @JSONSchema
const BeforeUserUpdateEventSchema = `
{
	"$id": "#BeforeUserUpdateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_user_update"] },
		"payload": { "$ref": "#UserUpdateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const AfterUserUpdateEventSchema = `
{
	"$id": "#AfterUserUpdateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_user_update"] },
		"payload": { "$ref": "#UserUpdateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const UserUpdateEventPayloadSchema = `
{
	"$id": "#UserUpdateEventPayload",
	"type": "object",
	"properties": {
		"reason": { "type": "string" },
		"metadata": { "type": "object" },
		"user": { "$ref": "#User" }
	}
}
`

func (UserUpdateEvent) BeforeEventType() Type {
	return BeforeUserUpdate
}

func (UserUpdateEvent) AfterEventType() Type {
	return AfterUserUpdate
}

func (event UserUpdateEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	// user object in this event is a snapshot before operation, so mutations are not applied
	newEvent := event
	if mutations.Metadata != nil {
		newEvent.Metadata = mutations.Metadata
	}
	return newEvent
}

func (event UserUpdateEvent) UserID() string {
	return event.User.ID
}
