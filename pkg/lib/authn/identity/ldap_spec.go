package identity

type LDAPSpec struct {
	ServerName           string                 `json:"server_name"`
	UserIDAttributeName  string                 `json:"user_id_attribute_name"`
	UserIDAttributeValue []byte                 `json:"user_id_attribute_value"`
	Claims               map[string]interface{} `json:"claims,omitempty"`
	RawEntryJSON         map[string]interface{} `json:"raw_entry_json,omitempty"`
	LastLoginUserName    *string                `json:"last_login_username"`
}
