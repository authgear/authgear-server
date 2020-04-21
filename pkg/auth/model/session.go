package model

import (
	"time"
)

// Session is the API model of user sessions
type Session struct {
	ID string `json:"id"`

	IdentityType   string                 `json:"identity_type"`
	IdentityClaims map[string]interface{} `json:"identity_claims"`
	ACR            string                 `json:"acr,omitempty"`
	AMR            []string               `json:"amr,omitempty"`

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
		"identity_type": { "type": "string" },
		"identity_claims": { "type": "object" },
		"acr": { "type": "string" },
		"amr": { "type": "array", "items": { "type": "string" } },
		"created_at": { "type": "string" },
		"last_accessed_at": { "type": "string" },
		"created_by_ip": { "type": "string" },
		"last_accessed_by_ip": { "type": "string" },
		"user_agent": { "$ref": "#UserAgent" }
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
