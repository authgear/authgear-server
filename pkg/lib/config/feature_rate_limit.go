package config

var _ = FeatureConfigSchema.Add("RateLimitBucketConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"size": { "type": "integer", "minimum": 0 },
		"reset_period": { "type": "integer", "minimum": 0 }
	}
}
`)

var _ = FeatureConfigSchema.Add("RateLimitFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"disabled": { "type": "boolean" },
		"sms": { "$ref": "#/$defs/RateLimitBucketConfig" }
	}
}
`)
