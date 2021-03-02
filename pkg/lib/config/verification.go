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
		"claims": { "$ref": "#/$defs/VerificationClaimsConfig" },
		"criteria": { "$ref": "#/$defs/VerificationCriteria" },
		"code_expiry_seconds": { "$ref": "#/$defs/DurationSeconds", "minimum": 60 }
	}
}
`)

type VerificationConfig struct {
	Claims     *VerificationClaimsConfig `json:"claims,omitempty"`
	Criteria   VerificationCriteria      `json:"criteria,omitempty"`
	CodeExpiry DurationSeconds           `json:"code_expiry_seconds,omitempty"`
}

func (c *VerificationConfig) SetDefaults() {
	if c.Criteria == "" {
		c.Criteria = VerificationCriteriaAny
	}
	if c.CodeExpiry == 0 {
		c.CodeExpiry = DurationSeconds(3600)
	}
}

var _ = Schema.Add("VerificationClaimsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": { "$ref": "#/$defs/VerificationClaimConfig" },
		"phone_number": { "$ref": "#/$defs/VerificationClaimConfig" }
	}
}
`)

type VerificationClaimsConfig struct {
	Email       *VerificationClaimConfig `json:"email,omitempty"`
	PhoneNumber *VerificationClaimConfig `json:"phone_number,omitempty"`
}

var _ = Schema.Add("VerificationClaimConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"required": { "type": "boolean" }
	}
}
`)

type VerificationClaimConfig struct {
	Enabled  *bool `json:"enabled,omitempty"`
	Required *bool `json:"required,omitempty"`
}

func (c *VerificationClaimConfig) SetDefaults() {
	if c.Enabled == nil {
		c.Enabled = newBool(true)
	}
	if c.Required == nil {
		c.Required = newBool(true)
	}
}

var _ = Schema.Add("VerificationOAuthClaimsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": { "$ref": "#/$defs/VerificationOAuthClaimConfig" }
	}
}
`)

type VerificationOAuthClaimsConfig struct {
	Email *VerificationOAuthClaimConfig `json:"email,omitempty"`
}

var _ = Schema.Add("VerificationOAuthClaimConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"assume_verified": { "type": "boolean" }
	}
}
`)

type VerificationOAuthClaimConfig struct {
	AssumeVerified *bool `json:"assume_verified,omitempty"`
}

func (c *VerificationOAuthClaimConfig) SetDefaults() {
	if c.AssumeVerified == nil {
		c.AssumeVerified = newBool(true)
	}
}
