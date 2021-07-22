package config

var _ = FeatureConfigSchema.Add("HookFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"blocking_handler": { "$ref": "#/$defs/BlockingHandlerFeatureConfig" },
		"non_blocking_handler": { "$ref": "#/$defs/NonBlockingHandlerFeatureConfig" }
	}
}
`)

type HookFeatureConfig struct {
	BlockingHandler    *BlockingHandlerFeatureConfig    `json:"blocking_handler,omitempty"`
	NonBlockingHandler *NonBlockingHandlerFeatureConfig `json:"non_blocking_handler,omitempty"`
}

var _ = FeatureConfigSchema.Add("BlockingHandlerFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" }
	}
}
`)

type BlockingHandlerFeatureConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}

func (c *BlockingHandlerFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
}

var _ = FeatureConfigSchema.Add("NonBlockingHandlerFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": { "type": "integer" }
	}
}
`)

type NonBlockingHandlerFeatureConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}

func (c *NonBlockingHandlerFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(99)
	}
}
