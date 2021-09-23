package config

var _ = Schema.Add("UserProfileConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"standard_attributes": { "$ref": "#/$defs/StandardAttributesConfig" }
	}
}
`)

type UserProfileConfig struct {
	StandardAttributes *StandardAttributesConfig `json:"standard_attributes,omitempty"`
}

var _ = Schema.Add("StandardAttributesConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"population": { "$ref": "#/$defs/StandardAttributesPopulationConfig" }
	}
}
`)

type StandardAttributesConfig struct {
	Population *StandardAttributesPopulationConfig `json:"population,omitempty"`
}

var _ = Schema.Add("StandardAttributesPopulationConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"strategy": {
			"type": "string",
			"enum": ["none", "on_signup"]
		}
	}
}
`)

type StandardAttributesPopulationStrategy string

const (
	StandardAttributesPopulationStrategyDefault  StandardAttributesPopulationStrategy = ""
	StandardAttributesPopulationStrategyNone     StandardAttributesPopulationStrategy = "none"
	StandardAttributesPopulationStrategyOnSignup StandardAttributesPopulationStrategy = "on_signup"
)

type StandardAttributesPopulationConfig struct {
	Strategy StandardAttributesPopulationStrategy `json:"strategy,omitempty"`
}

func (c *StandardAttributesPopulationConfig) SetDefaults() {
	if c.Strategy == StandardAttributesPopulationStrategyDefault {
		c.Strategy = StandardAttributesPopulationStrategyOnSignup
	}
}
