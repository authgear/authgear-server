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
		"code_expiry_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"sms": { "$ref": "#/$defs/VerificationSMSConfig" },
		"email": { "$ref": "#/$defs/VerificationEmailConfig" }
	}
}
`)

type VerificationConfig struct {
	Claims     *VerificationClaimsConfig `json:"claims,omitempty"`
	Criteria   VerificationCriteria      `json:"criteria,omitempty"`
	CodeExpiry DurationSeconds           `json:"code_expiry_seconds,omitempty"`
	SMS        *VerificationSMSConfig    `json:"sms,omitempty"`
	Email      *VerificationEmailConfig  `json:"email,omitempty"`
}

func (c *VerificationConfig) SetDefaults() {
	if c.Criteria == "" {
		c.Criteria = VerificationCriteriaAny
	}
	if c.CodeExpiry == 0 {
		c.CodeExpiry = DurationSeconds(3600)
	}
}

var _ = Schema.Add("VerificationSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"message": { "$ref": "#/$defs/SMSMessageConfig" }
	}
}
`)

type VerificationSMSConfig struct {
	Message SMSMessageConfig `json:"message,omitempty"`
}

var _ = Schema.Add("VerificationEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"message": { "$ref": "#/$defs/EmailMessageConfig" }
	}
}
`)

type VerificationEmailConfig struct {
	Message EmailMessageConfig `json:"message,omitempty"`
}

func (c *VerificationEmailConfig) SetDefaults() {
	if c.Message["subject"] == "" {
		c.Message["subject"] = "Email Verification Instruction"
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
