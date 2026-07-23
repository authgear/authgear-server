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

var _ MergeableFeatureConfig = &HookFeatureConfig{}

func (c *HookFeatureConfig) Merge(layer *FeatureConfig) MergeableFeatureConfig {
	if layer.Hook == nil {
		return c
	}

	merged := c
	if merged == nil {
		merged = &HookFeatureConfig{}
	}

	merged.BlockingHandler = merged.BlockingHandler.Merge(layer.Hook.BlockingHandler)
	merged.NonBlockingHandler = merged.NonBlockingHandler.Merge(layer.Hook.NonBlockingHandler)

	return merged
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
		c.Maximum = new(99)
	}
}

func (c *BlockingHandlerFeatureConfig) Merge(layer *BlockingHandlerFeatureConfig) *BlockingHandlerFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.Maximum != nil {
		c.Maximum = layer.Maximum
	}
	return c
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
		c.Maximum = new(99)
	}
}

func (c *NonBlockingHandlerFeatureConfig) Merge(layer *NonBlockingHandlerFeatureConfig) *NonBlockingHandlerFeatureConfig {
	if c == nil && layer == nil {
		return nil
	}
	if c == nil {
		return layer
	}
	if layer == nil {
		return c
	}
	if layer.Maximum != nil {
		c.Maximum = layer.Maximum
	}
	return c
}
