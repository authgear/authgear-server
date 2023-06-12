package config

type AuthenticationLockoutType string

const (
	AuthenticationLockoutTypePerUser      AuthenticationLockoutType = "per_user"
	AuthenticationLockoutTypePerUserPerIP AuthenticationLockoutType = "per_user_per_ip"
)

var _ = Schema.Add("AuthenticationLockoutType", `
{
	"type": "string",
	"enum": ["per_user", "per_user_per_ip"]
}
`)

var _ = Schema.Add("AuthenticationLockoutMethodConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"enabled": {
			"type": "boolean"
		}
	},
	"required": [
		"enabled"
	]
}
`)

type AuthenticationLockoutMethodConfig struct {
	Enabled bool `json:"enabled"`
}

var _ = Schema.Add("AuthenticationLockoutConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"max_attempts": {
			"type": "integer",
			"minimum": 0
		},
		"history_duration": {
			"$ref": "#/$defs/DurationString"
		},
		"minimum_duration": {
			"$ref": "#/$defs/DurationString"
		},
		"maximum_duration": {
			"$ref": "#/$defs/DurationString"
		},
		"backoff_factor": {
			"type": "number",
			"minimum": 1
		},
		"lockout_type": {
			"$ref": "#/$defs/AuthenticationLockoutType"
		},
		"password": {
			"$ref": "#/$defs/AuthenticationLockoutMethodConfig"
		},
		"totp": {
			"$ref": "#/$defs/AuthenticationLockoutMethodConfig"
		},
		"oob_otp": {
			"$ref": "#/$defs/AuthenticationLockoutMethodConfig"
		},
		"recovery_code": {
			"$ref": "#/$defs/AuthenticationLockoutMethodConfig"
		}
	},
	"allOf": [
		{
			"if": {
				"properties": {
					"max_attempts": {
						"type": "integer",
						"minimum": 1
					}
				},
				"required": ["max_attempts"]
			},
			"then": {
				"required": [
					"history_duration",
					"minimum_duration",
					"maximum_duration",
					"lockout_type"
				]
			}
		}
	]
}
`)

type AuthenticationLockoutConfig struct {
	MaxAttempts     int                                `json:"max_attempts,omitempty"`
	HistoryDuration DurationString                     `json:"history_duration,omitempty"`
	MinimumDuration DurationString                     `json:"minimum_duration,omitempty"`
	MaximumDuration DurationString                     `json:"maximum_duration,omitempty"`
	BackoffFactor   *float64                           `json:"backoff_factor,omitempty"`
	LockoutType     AuthenticationLockoutType          `json:"lockout_type,omitempty"`
	Password        *AuthenticationLockoutMethodConfig `json:"password,omitempty"`
	Totp            *AuthenticationLockoutMethodConfig `json:"totp,omitempty"`
	OOBOTP          *AuthenticationLockoutMethodConfig `json:"oob_otp,omitempty"`
	RecoveryCode    *AuthenticationLockoutMethodConfig `json:"recovery_code,omitempty"`
}

func (c *AuthenticationLockoutConfig) IsEnabled() bool {
	return c != nil && c.MaxAttempts > 0
}

func (c *AuthenticationLockoutConfig) SetDefaults() {
	if c.IsEnabled() {
		one := 1.0
		c.BackoffFactor = &one
	}
}
