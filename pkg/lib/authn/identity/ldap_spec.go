package identity

type LDAPSpec struct {
	ServerName           string                 `json:"server_name"`
	UserIDAttributeOID   string                 `json:"user_id_attribute_oid"`
	UserIDAttributeValue string                 `json:"user_id_attribute_value"`
	Claims               map[string]interface{} `json:"claims,omitempty"`
	RawEntryJSON         map[string]interface{} `json:"raw_entry_json,omitempty"`
}
