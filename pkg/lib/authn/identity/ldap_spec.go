package identity

type LDAPSpec struct {
	ServerName           string                 `json:"server_name"`
	UserIDAttribute      string                 `json:"user_id_attribute"`
	UserIDAttributeValue string                 `json:"user_id_attribute_value"`
	RawEntryJSON         map[string]interface{} `json:"raw_entry_json,omitempty"`
}
