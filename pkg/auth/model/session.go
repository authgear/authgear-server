package model

import (
	"time"
)

// Session is the API model of user sessions
type Session struct {
	ID string `json:"id"`

	IdentityID        string    `json:"identity_id"`
	IdentityType      string    `json:"identity_type"`
	IdentityUpdatedAt time.Time `json:"identity_updated_at"`

	AuthenticatorID         string     `json:"authenticator_id,omitempty"`
	AuthenticatorType       string     `json:"authenticator_type,omitempty"`
	AuthenticatorOOBChannel string     `json:"authenticator_oob_channel,omitempty"`
	AuthenticatorUpdatedAt  *time.Time `json:"authenticator_updated_at,omitempty"`

	CreatedAt        time.Time `json:"created_at"`
	LastAccessedAt   time.Time `json:"last_accessed_at"`
	CreatedByIP      string    `json:"created_by_ip"`
	LastAccessedByIP string    `json:"last_accessed_by_ip"`
	UserAgent        UserAgent `json:"user_agent"`
}

// @JSONSchema
const SessionSchema = `
{
	"$id": "#Session",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"identity_id": { "type": "string" },
		"created_at": { "type": "string" },
		"last_accessed_at": { "type": "string" },
		"created_by_ip": { "type": "string" },
		"last_accessed_by_ip": { "type": "string" },
		"user_agent": { "$ref": "#UserAgent" },
		"name": { "type": "string" },
		"data": { "type": "object" }
	}
}
`

// @JSONSchema
const SessionResponseSchema = `
{
	"$id": "#SessionResponse",
	"type": "object",
	"properties": {
		"result": { "$ref": "#Session" }
	}
}
`
