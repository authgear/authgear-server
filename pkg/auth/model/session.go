package model

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/model"
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

	CreatedAt        time.Time        `json:"created_at"`
	LastAccessedAt   time.Time        `json:"last_accessed_at"`
	CreatedByIP      string           `json:"created_by_ip"`
	LastAccessedByIP string           `json:"last_accessed_by_ip"`
	UserAgent        SessionUserAgent `json:"user_agent"`
}

type SessionModeler interface {
	AuthnAttrs() *authn.Attrs
	ToAPIModel() *Session
}

// SessionUserAgent is the API model of user agent of session
type SessionUserAgent model.UserAgent

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
		"user_agent": { "$ref": "#SessionUserAgent" },
		"name": { "type": "string" },
		"data": { "type": "object" }
	}
}
`

// @JSONSchema
const SessionUserAgentSchema = `
{
	"$id": "#SessionUserAgent",
	"type": "object",
	"properties": {
		"raw": { "type": "string" },
		"name": { "type": "string" },
		"version": { "type": "string" },
		"os": { "type": "string" },
		"os_version": { "type": "string" },
		"device_name": { "type": "string" },
		"device_model": { "type": "string" }
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
