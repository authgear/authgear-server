package config

var _ = Schema.Add("OTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"ratelimit": { "$ref": "#/$defs/OTPRatelimitConfig" }
	}
}
`)

type OTPConfig struct {
	Ratelimit *OTPRatelimitConfig `json:"ratelimit,omitempty"`
}

var _ = Schema.Add("OTPRatelimitConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"failed_attempt": { "$ref": "#/$defs/OTPFailedAttemptConfig" }
	}
}
`)

type OTPRatelimitConfig struct {
	FailedAttempt *OTPFailedAttemptConfig `json:"failed_attempt,omitempty"`
}

var _ = Schema.Add("OTPFailedAttemptConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"size": {
			"type": "integer",
			"minimum": 1
		},
		"reset_period": {
			"$ref": "#/$defs/DurationString",
			"format": "x_duration_string"
		}
	}
}
`)

type OTPFailedAttemptConfig struct {
	Size        int            `json:"size,omitempty"`
	ResetPeriod DurationString `json:"reset_period,omitempty"`
}

func (c *OTPFailedAttemptConfig) SetDefaults() {
	if c.Size == 0 {
		c.Size = 5
	}
	if c.ResetPeriod == "" {
		c.ResetPeriod = "20m"
	}
}
