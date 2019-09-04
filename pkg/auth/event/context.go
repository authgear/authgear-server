package event

import (
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

// @JSONSchema
const ContextSchema = `
{
	"$id": "#EventContext",
	"type": "object",
	"properties": {
		"timestamp": { "type": "integer" },
		"request_id": { "type": "string" },
		"user_id": { "type": "string" },
		"identity_id": { "type": "string" },
		"session": { "$ref": "#Session" }
	}
}
`

type Context struct {
	Timestamp   int64          `json:"timestamp"`
	RequestID   *string        `json:"request_id"`
	UserID      *string        `json:"user_id"`
	PrincipalID *string        `json:"identity_id"`
	Session     *model.Session `json:"session"`
}
