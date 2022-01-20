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
	TriggeredBySystem       TriggeredByType = "system"
)

type Context struct {
	Timestamp          int64           `json:"timestamp"`
	UserID             *string         `json:"user_id"`
	TriggeredBy        TriggeredByType `json:"triggered_by"`
	PreferredLanguages []string        `json:"preferred_languages"`
	Language           string          `json:"language"`

	OAuth     *OAuthContext `json:"oauth,omitempty"`
	IPAddress string        `json:"ip_address,omitempty"`
	UserAgent string        `json:"user_agent,omitempty"`
	ClientID  string        `json:"client_id,omitempty"`
}

type OAuthContext struct {
	State string `json:"state,omitempty"`
}
