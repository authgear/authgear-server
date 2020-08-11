package model

import (
	"time"
)

type User struct {
	ID          string                 `json:"id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	LastLoginAt *time.Time             `json:"last_login_at,omitempty"`
	IsAnonymous bool                   `json:"is_anonymous"`
	IsVerified  bool                   `json:"is_verified"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// @JSONSchema
const UserSchema = `
{
	"$id": "#User",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"created_at": { "type": "string" },
		"last_login_at": { "type": "string" },
		"is_anonymous": { "type": "boolean" },
		"metadata": { "type": "object" }
	}
}
`
