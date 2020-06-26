package config

var _ = Schema.Add("AppMetadata", `
{
	"type": "object",
	"additionalProperties": false,
	"patternProperties": {
		"^logo_uri(#.+)?$": { "type": "string", "format": "uri" },
		"^app_name(#.+)?$": { "type": "string" }
	}
}
`)

type AppMetadata map[string]interface{}
