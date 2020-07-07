package config

var _ = Schema.Add("VerificationCriteria", `
{
	"type": "string",
	"enum": ["any", "all"]
}
`)

type VerificationCriteria string

const (
	VerificationCriteriaAny VerificationCriteria = "any"
	VerificationCriteriaAll VerificationCriteria = "all"
)

var _ = Schema.Add("VerificationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"criteria": { "$ref": "#/$defs/VerificationCriteria" },
		"login_id_keys": { "type": "array", "items": { "type": "string" } }
	}
}
`)

type VerificationConfig struct {
	Criteria    VerificationCriteria `json:"criteria,omitempty"`
	LoginIDKeys []string             `json:"login_id_keys,omitempty"`
}

func (c *VerificationConfig) SetDefaults() {
	if c.Criteria == "" {
		c.Criteria = VerificationCriteriaAny
	}
	if c.LoginIDKeys == nil {
		c.LoginIDKeys = []string{"email", "phone"}
	}
}
