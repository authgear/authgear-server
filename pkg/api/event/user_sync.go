package event

import "github.com/authgear/authgear-server/pkg/api/model"

const (
	UserSync Type = "user_sync"
)

/*
	@Callback
		@Operation POST /user_sync - Synchronize user information
			User information should be synchronized.
			@RequestBody
				@JSONSchema {UserSyncEvent}
			@Response 200 {EmptyResponse}
*/
type UserSyncEvent struct {
	User model.User `json:"user"`
}

// @JSONSchema
const UserSyncEventSchema = `
{
	"$id": "#UserSyncEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["user_sync"] },
		"payload": { "$ref": "#UserSyncEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// @JSONSchema
const UserSyncEventPayloadSchema = `
{
	"$id": "#UserSyncEventPayload",
	"type": "object",
	"properties": {
		"user": { "$ref": "#User" }
	}
}
`

func (e *UserSyncEvent) EventType() Type {
	return UserSync
}

func (e *UserSyncEvent) UserID() string {
	return e.User.ID
}
