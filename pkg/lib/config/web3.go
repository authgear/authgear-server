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

type Deprecated_NFTConfig struct {
	Collections []string `json:"collections,omitempty"`
}

var _ = Schema.Add("SIWEConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"networks": { 
			"type": "array",
			"items": {
				"type": "string",
				"format": "x_web3_network_id",
				"minLength": 1
			},
			"uniqueItems": true
		}
	}
}
`)

type Deprecated_SIWEConfig struct {
	Networks []string `json:"networks,omitempty"`
}

var _ = Schema.Add("Web3Config", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"siwe": { "$ref": "#/$defs/SIWEConfig" },
		"nft": { "$ref": "#/$defs/NFTConfig" }
	}
}
`)

type Deprecated_Web3Config struct {
	SIWE *Deprecated_SIWEConfig `json:"siwe,omitempty"`
	NFT  *Deprecated_NFTConfig  `json:"nft,omitempty"`
}
