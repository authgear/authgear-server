package config

var _ = Schema.Add("NFTConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"collections": {
			"type": "array",
			"items": { 
				"type": "string"
			},
			"uniqueItems": true
		}
	}
}
`)

type NFTConfig struct {
	Collections []string `json:"collections,omitempty"`
}
