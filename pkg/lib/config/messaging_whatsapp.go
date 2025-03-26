package config

var _ = Schema.Add("WhatsappAPIType", `
{
	"type": "string",
	"enum": ["on-premises", "cloud-api"]
}
`)

type WhatsappAPIType string

const (
	WhatsappAPITypeOnPremises WhatsappAPIType = "on-premises"
	WhatsappAPITypeCloudAPI   WhatsappAPIType = "cloud-api"
)

var _ = Schema.Add("WhatsappConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"api_type": { "$ref": "#/$defs/WhatsappAPIType" }
	}
}
`)

type WhatsappConfig struct {
	APIType WhatsappAPIType `json:"api_type,omitempty"`
}

func (c *WhatsappConfig) SetDefaults() {
	if string(c.APIType) == "" {
		c.APIType = WhatsappAPITypeOnPremises
	}
}
