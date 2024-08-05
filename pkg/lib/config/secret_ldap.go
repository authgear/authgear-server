package config

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
		"name": { "type": "string", "minLength": 1 },
		"dn": { "type": "string", "format": "ldap_dn", "minLength": 1 },
		"password": { "type": "string", "minLength": 1 }
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
