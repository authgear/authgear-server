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
	// TriggeredByTypeUser means the event originates from a end-user facing UI.
	TriggeredByTypeUser TriggeredByType = "user"
	// TriggeredByTypeAdminAPI means the event originates from the Admin API.
	TriggeredByTypeAdminAPI TriggeredByType = "admin_api"
	// TriggeredBySystem means the event originates from a background job.
	TriggeredBySystem TriggeredByType = "system"
	// TriggeredByPortal means the event originates from the management portal.
	TriggeredByPortal TriggeredByType = "portal"
)

type Context struct {
	Timestamp          int64           `json:"timestamp"`
	UserID             *string         `json:"user_id"`
	TriggeredBy        TriggeredByType `json:"triggered_by"`
	AuditContext       AuditContext    `json:"audit_context"`
	PreferredLanguages []string        `json:"preferred_languages"`
	Language           string          `json:"language"`

	OAuth           *OAuthContext `json:"oauth,omitempty"`
	IPAddress       string        `json:"ip_address,omitempty"`
	GeoLocationCode *string       `json:"geo_location_code"`
	UserAgent       string        `json:"user_agent,omitempty"`
	AppID           string        `json:"app_id,omitempty"`
	ClientID        string        `json:"client_id,omitempty"`
	TrackingID      string        `json:"tracking_id,omitempty"`
}

type OAuthContext struct {
	State  string `json:"state,omitempty"`
	XState string `json:"x_state,omitempty"`
}

type AuditContext map[string]any

func NewAuditContext(httpURL string, info map[string]any) AuditContext {
	auditCtx := AuditContext{}
	for k, v := range info {
		auditCtx[k] = v
	}
	auditCtx["http_url"] = httpURL
	return auditCtx
}
