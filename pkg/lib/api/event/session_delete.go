package event

import "github.com/authgear/authgear-server/pkg/auth/model"

const (
	BeforeSessionDelete Type = "before_session_delete"
	AfterSessionDelete  Type = "after_session_delete"
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
	Reason  string        `json:"reason"`
	User    model.User    `json:"user"`
	Session model.Session `json:"session"`
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
		"session": { "$ref": "#Session" }
	}
}
`

func (e *SessionDeleteEvent) BeforeEventType() Type {
	return BeforeSessionDelete
}

func (e *SessionDeleteEvent) AfterEventType() Type {
	return AfterSessionDelete
}

func (e *SessionDeleteEvent) UserID() string {
	return e.User.ID
}
