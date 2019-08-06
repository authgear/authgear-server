package event

// @JSONSchema
const ContextSchema = `
{
	"$id": "#EventContext",
	"type": "object",
	"properties": {
		"timestamp": { "type": "integer" },
		"request_id": { "type": "string" },
		"user_id": { "type": "string" },
		"identity_id": { "type": "string" }
	}
}
`

type Context struct {
	Timestamp   int64   `json:"timestamp"`
	RequestID   *string `json:"request_id"`
	UserID      *string `json:"user_id"`
	PrincipalID *string `json:"identity_id"`
}
