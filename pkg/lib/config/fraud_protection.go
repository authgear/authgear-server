package config

var _ = Schema.Add("FraudProtectionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"warnings": {
			"type": "array",
			"items": { "$ref": "#/$defs/FraudProtectionWarning" }
		},
		"decision": { "$ref": "#/$defs/FraudProtectionDecision" }
	}
}
`)

var _ = Schema.Add("FraudProtectionWarning", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["type"],
	"properties": {
		"type": { "$ref": "#/$defs/FraudProtectionWarningType" }
	}
}
`)

var _ = Schema.Add("FraudProtectionWarningType", `
{
	"type": "string",
	"enum": [
		"SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED",
		"SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED",
		"SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED",
		"SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED",
		"SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED"
	]
}
`)

var _ = Schema.Add("FraudProtectionDecision", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"always_allow": { "$ref": "#/$defs/FraudProtectionAlwaysAllow" },
		"action": { "$ref": "#/$defs/FraudProtectionDecisionAction" }
	}
}
`)

var _ = Schema.Add("FraudProtectionAlwaysAllow", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ip_address": { "$ref": "#/$defs/FraudProtectionIPAlwaysAllow" },
		"phone_number": { "$ref": "#/$defs/FraudProtectionPhoneNumberAlwaysAllow" }
	}
}
`)

var _ = Schema.Add("FraudProtectionIPAlwaysAllow", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"cidrs": {
			"type": "array",
			"items": { "type": "string", "format": "x_cidr" }
		},
		"geo_location_codes": {
			"type": "array",
			"items": { "type": "string", "minLength": 2, "maxLength": 2 }
		}
	}
}
`)

var _ = Schema.Add("FraudProtectionPhoneNumberAlwaysAllow", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"geo_location_codes": {
			"type": "array",
			"items": { "type": "string", "minLength": 2, "maxLength": 2 }
		},
		"regex": {
			"type": "array",
			"items": { "type": "string", "format": "x_re2_regex" }
		}
	}
}
`)

var _ = Schema.Add("FraudProtectionDecisionAction", `
{
	"type": "string",
	"enum": ["record_only", "deny_if_any_warning"]
}
`)

type FraudProtectionConfig struct {
	Enabled  *bool                       `json:"enabled,omitempty"`
	Warnings []*FraudProtectionWarning   `json:"warnings,omitempty"`
	Decision *FraudProtectionDecision    `json:"decision,omitempty"`
}

func (c *FraudProtectionConfig) SetDefaults() {
	if c.Enabled == nil {
		c.Enabled = newBool(true)
	}
	if c.Warnings == nil {
		c.Warnings = []*FraudProtectionWarning{
			{Type: FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily},
			{Type: FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily},
			{Type: FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly},
			{Type: FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily},
			{Type: FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly},
		}
	}
}

type FraudProtectionWarning struct {
	Type FraudProtectionWarningType `json:"type"`
}

type FraudProtectionWarningType string

const (
	FraudProtectionWarningTypeSMSPhoneCountriesByIPDaily            FraudProtectionWarningType = "SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED"
	FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryDaily  FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED"
	FraudProtectionWarningTypeSMSUnverifiedOTPsByPhoneCountryHourly FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED"
	FraudProtectionWarningTypeSMSUnverifiedOTPsByIPDaily            FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED"
	FraudProtectionWarningTypeSMSUnverifiedOTPsByIPHourly           FraudProtectionWarningType = "SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED"
)

type FraudProtectionDecision struct {
	AlwaysAllow *FraudProtectionAlwaysAllow   `json:"always_allow,omitempty" nullable:"true"`
	Action      FraudProtectionDecisionAction `json:"action,omitempty"`
}

func (c *FraudProtectionDecision) SetDefaults() {
	if c.Action == "" {
		c.Action = FraudProtectionDecisionActionRecordOnly
	}
}

type FraudProtectionAlwaysAllow struct {
	IPAddress   *FraudProtectionIPAlwaysAllow          `json:"ip_address,omitempty" nullable:"true"`
	PhoneNumber *FraudProtectionPhoneNumberAlwaysAllow `json:"phone_number,omitempty" nullable:"true"`
}

type FraudProtectionIPAlwaysAllow struct {
	CIDRs            []string `json:"cidrs,omitempty"`
	GeoLocationCodes []string `json:"geo_location_codes,omitempty"`
}

type FraudProtectionPhoneNumberAlwaysAllow struct {
	GeoLocationCodes []string `json:"geo_location_codes,omitempty"`
	Regex            []string `json:"regex,omitempty"`
}

type FraudProtectionDecisionAction string

const (
	FraudProtectionDecisionActionRecordOnly       FraudProtectionDecisionAction = "record_only"
	FraudProtectionDecisionActionDenyIfAnyWarning FraudProtectionDecisionAction = "deny_if_any_warning"
)
