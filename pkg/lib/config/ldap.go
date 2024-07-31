package config

import (
	"net/url"
)

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
	Servers []LDAPServerConfig `json:"servers,omitempty"`
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
	Name                             string   `json:"name,omitempty"`
	URL                              *url.URL `json:"url,omitempty"`
	BaseDN                           string   `json:"base_dn,omitempty"`
	SearchFilterTemplate             string   `json:"search_filter_template,omitempty"`
	UserUniqueIdentifierAttributeOID string   `json:"user_id_attribute_oid,omitempty"`
}

var _ = SecretConfigSchema.Add("LDAPServerUserCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"items": { "type": "array", "items": { "$ref": "#/$defs/LDAPServerUserCredentialsItem" } }
	}
}
`)

type LDAPServerUserCredentials struct {
	Items []LDAPServerUserCredentialsItem `json:"items,omitempty"`
}

var _ = SecretConfigSchema.Add("LDAPServerUserCredentialsItem", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string" },
		"dn": { "type": "string", "format": "ldap_dn" },
		"password": { "type": "string" }
	},
	"required": ["name"],
	"dependentRequired": {
		"dn": ["password"],
		"password": ["dn"]
	}
}
`)

type LDAPServerUserCredentialsItem struct {
	Name     string `json:"name,omitempty"`
	DN       string `json:"dn,omitempty"`
	Password string `json:"password,omitempty"`
}

func (c *LDAPServerUserCredentials) SensitiveStrings() []string {
	var out []string
	for _, item := range c.Items {
		out = append(out, item.SensitiveStrings()...)
	}
	return out
}

func (c *LDAPServerUserCredentialsItem) SensitiveStrings() []string {
	return []string{c.Password}
}
