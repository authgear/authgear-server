package event

import "github.com/skygeario/skygear-server/pkg/auth/model"

const (
	BeforeSessionDelete Type = "before_session_delete"
	AfterSessionDelete  Type = "after_session_delete"
)

type SessionDeleteReason string

const (
	SessionDeleteReasonLogout = "logout"
)

/*
	@Callback
		@Operation POST /before_session_delete - Before session deletion
			A session is about to be created.
			@RequestBody
				@JSONSchema {BeforeSessionDeleteEvent}
			@Response 200 {HookResponse}

		@Operation POST /after_session_delete - After session deletion
			A session is created.
			@RequestBody
				@JSONSchema {AfterSessionDeleteEvent}
			@Response 200 {EmptyResponse}
*/
type SessionDeleteEvent struct {
	Reason   SessionDeleteReason `json:"reason"`
	User     model.User          `json:"user"`
	Identity model.Identity      `json:"identity"`
}

// @JSONSchema
const BeforeSessionDeleteEventSchema = `
{
	"$id": "#BeforeSessionDeleteEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_session_delete"] },
		"payload": { "$ref": "#SessionDeleteEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const AfterSessionDeleteEventSchema = `
{
	"$id": "#AfterSessionDeleteEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_session_delete"] },
		"payload": { "$ref": "#SessionDeleteEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const SessionDeleteEventPayloadSchema = `
{
	"$id": "#SessionDeleteEventPayload",
	"type": "object",
	"properties": {
		"reason": { "type": "string" },
		"user": { "$ref": "#User" },
		"identity": { "$ref": "#Identity" }
	}
}
`

func (SessionDeleteEvent) BeforeEventType() Type {
	return BeforeSessionDelete
}

func (SessionDeleteEvent) AfterEventType() Type {
	return AfterSessionDelete
}

func (event SessionDeleteEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	user := event.User
	mutations.ApplyToUser(&user)
	return SessionDeleteEvent{
		Reason:   event.Reason,
		User:     user,
		Identity: event.Identity,
	}
}

func (event SessionDeleteEvent) UserID() string {
	return event.User.ID
}
