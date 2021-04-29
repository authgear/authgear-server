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

type TriggeredByType string

const (
	TriggeredByTypeUser     TriggeredByType = "user"
	TriggeredByTypeAdminAPI TriggeredByType = "admin_api"
)

type Context struct {
	Timestamp          int64           `json:"timestamp"`
	UserID             *string         `json:"user_id"`
	PreferredLanguages []string        `json:"preferred_languages"`
	Language           string          `json:"language"`
	TriggeredBy        TriggeredByType `json:"triggered_by"`
	OAuth              *OAuthContext   `json:"oauth,omitempty"`
}

type OAuthContext struct {
	State string `json:"state,omitempty"`
}
