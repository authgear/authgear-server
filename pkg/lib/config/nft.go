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
