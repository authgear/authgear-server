package config

import "github.com/authgear/authgear-server/pkg/core/auth/metadata"

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
		"criteria": { "$ref": "#/$defs/VerificationCriteria" }
	}
}
`)

type VerificationConfig struct {
	Criteria VerificationCriteria `json:"criteria,omitempty"`
}

func (c *VerificationConfig) SetDefaults() {
	if c.Criteria == "" {
		c.Criteria = VerificationCriteriaAny
	}
}

var _ = Schema.Add("VerificationLoginIDKeyConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"required": { "type": "boolean" }
	}
}
`)

type VerificationLoginIDKeyConfig struct {
	Enabled  *bool `json:"enabled,omitempty"`
	Required *bool `json:"required,omitempty"`
}

func (c *VerificationLoginIDKeyConfig) SetDefaults(keyType LoginIDKeyType) {
	isVerifiableType := false
	if m, ok := keyType.MetadataKey(); ok && (m == metadata.Email || m == metadata.Phone) {
		isVerifiableType = true
	}

	if c.Enabled == nil {
		c.Enabled = newBool(isVerifiableType)
	}
	if c.Required == nil {
		c.Required = newBool(isVerifiableType)
	}
}
