package config

var _ = Schema.Add("LDAPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"servers": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/LDAPServerConfig"
			},
			"minItems": 1
		}
	}
}
`)

type LDAPConfig struct {
	Servers []*LDAPServerConfig `json:"servers,omitempty"`
}

var _ = Schema.Add("LDAPServerConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["name", "url", "base_dn", "search_filter_template", "user_id_attribute_oid"],
	"properties": {
		"name": { "type": "string", "minLength": 1 },
		"url": { "type": "string", "format": "ldap_url" },
		"base_dn": { "type": "string", "format": "ldap_dn" },
		"search_filter_template": { "type": "string", "format": "ldap_search_filter_template" },
		"user_id_attribute_oid": { "type": "string", "format": "ldap_oid" }
	}
}
`)

type LDAPServerConfig struct {
	Name                             string `json:"name,omitempty"`
	URL                              string `json:"url,omitempty"`
	BaseDN                           string `json:"base_dn,omitempty"`
	SearchFilterTemplate             string `json:"search_filter_template,omitempty"`
	UserUniqueIdentifierAttributeOID string `json:"user_id_attribute_oid,omitempty"`
}
