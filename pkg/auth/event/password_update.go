package event

import "github.com/authgear/authgear-server/pkg/auth/model"

const (
	BeforePasswordUpdate Type = "before_password_update"
	AfterPasswordUpdate  Type = "after_password_udpate"
)

type PasswordUpdateReason string

const (
	PasswordUpdateReasonChangePassword = "change_password"
	PasswordUpdateReasonResetPassword  = "reset_password"
	PasswordUpdateReasonAdministrative = "administrative"
)

/*
	@Callback
		@Operation POST /before_password_update - Before password update
			The password of a user is about to be updated.
			@RequestBody
				@JSONSchema {BeforePasswordUpdateEvent}
			@Response 200 {HookResponse}

		@Operation POST /after_password_udpate - After password update
			The password of a user is created.
			@RequestBody
				@JSONSchema {AfterPasswordUpdateEvent}
			@Response 200 {EmptyResponse}
*/
type PasswordUpdateEvent struct {
	Reason PasswordUpdateReason `json:"reason"`
	User   model.User           `json:"user"`
}

// nolint: gosec
// @JSONSchema
const BeforePasswordUpdateEventSchema = `
{
	"$id": "#BeforePasswordUpdateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["before_password_update"] },
		"payload": { "$ref": "#PasswordUpdateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// nolint: gosec
// @JSONSchema
const AfterPasswordUpdateEventSchema = `
{
	"$id": "#AfterPasswordUpdateEvent",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"seq": { "type": "integer" },
		"type": { "type": "string", "enum": ["after_password_update"] },
		"payload": { "$ref": "#PasswordUpdateEventPayload" },
		"context": { "$ref": "#EventContext" }
	}
}
`

// nolint: gosec
// @JSONSchema
const PasswordUpdateEventPayloadSchema = `
{
	"$id": "#PasswordUpdateEventPayload",
	"type": "object",
	"properties": {
		"reason": { "type": "string" },
		"user": { "$ref": "#User" }
	}
}
`

func (PasswordUpdateEvent) BeforeEventType() Type {
	return BeforePasswordUpdate
}

func (PasswordUpdateEvent) AfterEventType() Type {
	return AfterPasswordUpdate
}

func (event PasswordUpdateEvent) WithMutationsApplied(mutations Mutations) UserAwarePayload {
	// user object in this event is a snapshot before operation, so mutations are not applied
	return event
}

func (event PasswordUpdateEvent) UserID() string {
	return event.User.ID
}
