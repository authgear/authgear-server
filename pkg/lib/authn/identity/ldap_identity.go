package identity

import "time"

type LDAP struct {
	ID                   string                 `json:"id"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	UserID               string                 `json:"user_id"`
	ServerName           string                 `json:"server_name"`
	UserIDAttribute      string                 `json:"user_id_attribute"`
	UserIDAttributeValue string                 `json:"user_id_attribute_value"`
	Claims               map[string]interface{} `json:"claims,omitempty"`
	RawEntryJSON         map[string]interface{} `json:"raw_entry_json,omitempty"`
}
