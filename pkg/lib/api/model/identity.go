package model

type Identity struct {
	Type   string                 `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}

// @JSONSchema
const IdentitySchema = `
{
	"$id": "#Identity",
	"type": "object",
	"properties": {
		"type": { "type": "string" },
		"claims": { "type": "object" }
	}
}
`
