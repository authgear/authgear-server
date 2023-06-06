package config

var _ = Schema.Add("AuthenticationLockoutConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"max_attempts": {
			"type": "integer"
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
			"type": "number"
		}
	},
	"required": ["max_attempts", "history_duration", "minimum_duration", "maximum_duration", "backoff_factor"]
}
`)

type AuthenticationLockoutConfig struct {
	MaxAttempts     int            `json:"max_attempts,omitempty"`
	HistoryDuration DurationString `json:"history_duration,omitempty"`
	MinimumDuration DurationString `json:"minimum_duration,omitempty"`
	MaximumDuration DurationString `json:"maximum_duration,omitempty"`
	BackoffFactor   float64        `json:"backoff_factor,omitempty"`
}
