package config

var _ = FeatureConfigSchema.Add("Web3FeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"nft": { "$ref": "#/$defs/Web3NFTFeatureConfig" }
	}
}	
`)

type Web3FeatureConfig struct {
	NFT *Web3NFTFeatureConfig `json:"nft,omitempty"`
}

var _ = FeatureConfigSchema.Add("Web3NFTFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"maximum": {"type": "integer"}
	}
}
`)

type Web3NFTFeatureConfig struct {
	Maximum *int `json:"maximum,omitempty"`
}

func (c *Web3NFTFeatureConfig) SetDefaults() {
	if c.Maximum == nil {
		c.Maximum = newInt(3)
	}
}
