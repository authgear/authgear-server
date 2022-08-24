package config

var _ = Schema.Add("NFTConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"collections": {
			"type": "array",
			"items": { 
				"type": "string",
				"format": "x_web3_contract_id",
				"minLength": 1
			},
			"uniqueItems": true
		}
	}
}
`)

type NFTConfig struct {
	Collections []string `json:"collections,omitempty"`
}

var _ = Schema.Add("Web3Config", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"nft": { "$ref": "#/$defs/NFTConfig" }
	}
}
`)

type Web3Config struct {
	NFT *NFTConfig `json:"nft,omitempty"`
}
