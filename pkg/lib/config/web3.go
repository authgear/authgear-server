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

type BlockchainType string

const (
	BlockchainTypeEthereum BlockchainType = "ethereum"
)

var _ = Schema.Add("BlockchainConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"blockchain": { "type": "string", "enum": ["ethereum"] },
		"network": { "type": "string" }
	},
	"required": ["blockchain", "network"],
	"allOf": [
		{
			"if": {
				"properties": {
					"blockchain": {
						"const": "ethereum"
					}
				},
				"required": ["blockchain"]
			},
			"then": {
				"properties": {
					"network": {
						"format": "x_web3_ethereum_chain_id"
					}
				},
				"required": ["network"]
			}
		}
	]
}
`)

type BlockchainConfig struct {
	Blockchain BlockchainType `json:"blockchain,omitempty"`
	Network    string         `json:"network,omitempty"`
}

func (b *BlockchainConfig) SetDefaults() {
	// Default ethereum blockchain
	if b.Blockchain == "" {
		b.Blockchain = BlockchainTypeEthereum
	}

	// Default ethereum mainnet
	if b.Network == "" {
		b.Network = "1"
	}
}

var _ = Schema.Add("Web3Config", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"blockchain": { "$ref": "#/$defs/BlockchainConfig" },
		"nft": { "$ref": "#/$defs/NFTConfig" }
	}
}
`)

type Web3Config struct {
	Blockchain *BlockchainConfig `json:"blockchain,omitempty"`
	NFT        *NFTConfig        `json:"nft,omitempty"`
}
