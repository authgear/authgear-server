package event

// @JSONSchema
const ContextSchema = `
{
	"$id": "#EventContext",
	"type": "object",
	"properties": {
		"timestamp": { "type": "integer" },
		"user_id": { "type": "string" }
	}
}
`

type Context struct {
	Timestamp          int64    `json:"timestamp"`
	UserID             *string  `json:"user_id"`
	PreferredLanguages []string `json:"preferred_languages"`
	Language           string   `json:"language"`
}
