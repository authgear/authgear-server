package config

var _ = Schema.Add("SMSGatewayConfigUseConfigFrom", `
{
	"type": "string",
	"enum": ["environment_variable", "authgear.secrets.yaml"]
}
`)

type SMSGatewayConfigUseConfigFrom string

const (
	SMSGatewayUseConfigFromEnvironmentVariable SMSGatewayConfigUseConfigFrom = "environment_variable"
	SMSGatewayUseConfigFromAuthgearSecretsYAML SMSGatewayConfigUseConfigFrom = "authgear.secrets.yaml"
)

var _ = Schema.Add("SMSGatewayConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"use_config_from": { "$ref": "#/$defs/SMSGatewayConfigUseConfigFrom" },
		"provider": { "$ref": "#/$defs/SMSProvider" }
	},
	"required": ["use_config_from"],
	"allOf": [
		{
			"if": {
				"properties": { "use_config_from": { "const": "authgear.secrets.yaml" } },
				"required": ["use_config_from"]
			},
			"then": { "required": ["provider"] }
		}
	]
}
`)

type SMSGatewayConfig struct {
	UseConfigFrom SMSGatewayConfigUseConfigFrom `json:"use_config_from,omitempty"`
	Provider      SMSProvider                   `json:"provider,omitempty"`
}
