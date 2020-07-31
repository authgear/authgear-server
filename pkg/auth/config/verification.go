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
		"code_expiry_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"sms": { "$ref": "#/$defs/VerificationSMSConfig" },
		"email": { "$ref": "#/$defs/VerificationEmailConfig" }
	}
}
`)

type VerificationConfig struct {
	Criteria   VerificationCriteria     `json:"criteria,omitempty"`
	CodeExpiry DurationSeconds          `json:"code_expiry_seconds,omitempty"`
	SMS        *VerificationSMSConfig   `json:"sms,omitempty"`
	Email      *VerificationEmailConfig `json:"email,omitempty"`
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
		"message": { "$ref": "#/$defs/SMSMessageConfig" },
		"code_format": { "$ref": "#/$defs/OTPFormat" }
	}
}
`)

var _ = Schema.Add("OTPFormat", `
{
	"type": "string",
	"enum": ["numeric", "complex"]
}
`)

type OTPFormat string

const (
	OTPFormatNumeric OTPFormat = "numeric"
	OTPFormatComplex OTPFormat = "complex"
)

type VerificationSMSConfig struct {
	Message    SMSMessageConfig `json:"message,omitempty"`
	CodeFormat OTPFormat        `json:"code_format,omitempty"`
}

func (c *VerificationSMSConfig) SetDefaults() {
	if c.CodeFormat == "" {
		c.CodeFormat = OTPFormatNumeric
	}
}

var _ = Schema.Add("VerificationEmailConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"message": { "$ref": "#/$defs/EmailMessageConfig" },
		"code_format": { "$ref": "#/$defs/OTPFormat" }
	}
}
`)

type VerificationEmailConfig struct {
	Message    EmailMessageConfig `json:"message,omitempty"`
	CodeFormat OTPFormat          `json:"code_format,omitempty"`
}

func (c *VerificationEmailConfig) SetDefaults() {
	if c.Message["subject"] == "" {
		c.Message["subject"] = "Email Verification Instruction"
	}
	if c.CodeFormat == "" {
		c.CodeFormat = OTPFormatComplex
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
	if keyType == LoginIDKeyTypeEmail || keyType == LoginIDKeyTypePhone {
		isVerifiableType = true
	}

	if c.Enabled == nil {
		c.Enabled = newBool(isVerifiableType)
	}
	if c.Required == nil {
		c.Required = newBool(isVerifiableType)
	}
}
