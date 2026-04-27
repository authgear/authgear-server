package config

var _ = Schema.Add("FraudProtectionConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": { "type": "boolean" },
		"sms": {
			"$ref": "#/$defs/FraudProtectionSMSConfig"
		},
		"warnings": {
			"type": "array",
			"items": { "$ref": "#/$defs/FraudProtectionWarning" }
		},
		"decision": { "$ref": "#/$defs/FraudProtectionDecision" }
	}
}
`)

var _ = Schema.Add("FraudProtectionSMSConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"unverified_otp_budget": {
			"$ref": "#/$defs/FraudProtectionSMSUnverifiedOTPBudgetConfig"
		}
	}
}
`)

var _ = Schema.Add("FraudProtectionSMSUnverifiedOTPBudgetConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"daily_ratio": {
			"type": "number",
			"minimum": 0,
			"maximum": 1
		},
		"hourly_ratio": {
			"type": "number",
			"minimum": 0,
			"maximum": 1
		},
		"by_phone_country": {
			"type": "array",
			"items": { "$ref": "#/$defs/FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig" }
		}
	}
}
`)

var _ = Schema.Add("FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"required": ["country_codes"],
	"properties": {
		"country_codes": {
			"type": "array",
			"items": {
				"type": "string",
				"minLength": 2,
				"maxLength": 2
			}
		},
		"daily_ratio": {
			"type": "number",
			"minimum": 0,
			"maximum": 1
		},
		"hourly_ratio": {
			"type": "number",
			"minimum": 0,
			"maximum": 1
		}
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
	Enabled  *bool                     `json:"enabled,omitempty"`
	SMS      *FraudProtectionSMSConfig `json:"sms,omitempty"`
	Warnings []*FraudProtectionWarning `json:"warnings,omitempty"`
	Decision *FraudProtectionDecision  `json:"decision,omitempty"`
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

type FraudProtectionSMSConfig struct {
	UnverifiedOTPBudget *FraudProtectionSMSUnverifiedOTPBudgetConfig `json:"unverified_otp_budget,omitempty"`
}

type FraudProtectionSMSUnverifiedOTPBudgetConfig struct {
	DailyRatio     *float64                                                     `json:"daily_ratio,omitempty"`
	HourlyRatio    *float64                                                     `json:"hourly_ratio,omitempty"`
	ByPhoneCountry []*FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig `json:"by_phone_country,omitempty"`
}

func (c *FraudProtectionSMSUnverifiedOTPBudgetConfig) SetDefaults() {
	if c.DailyRatio == nil {
		c.DailyRatio = newFloat64(0.3)
	}
	if c.HourlyRatio == nil {
		c.HourlyRatio = newFloat64(0.2)
	}
}

type FraudProtectionSMSUnverifiedOTPBudgetByPhoneCountryConfig struct {
	CountryCodes []string `json:"country_codes,omitempty"`
	DailyRatio   *float64 `json:"daily_ratio,omitempty" nullable:"true"`
	HourlyRatio  *float64 `json:"hourly_ratio,omitempty" nullable:"true"`
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
