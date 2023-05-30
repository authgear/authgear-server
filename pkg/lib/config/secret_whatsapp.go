package config

var _ = SecretConfigSchema.Add("WhatsappOnPremisesCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"api_endpoint": { "type": "string", "minLength": 1 },
		"username": { "type": "string", "minLength": 1 },
		"password": { "type": "string", "minLength": 1 }
	},
	"required": ["api_endpoint", "username", "password"]
}
`)

type WhatsappOnPremisesCredentials struct {
	APIEndpoint string `json:"api_endpoint"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func (c *WhatsappOnPremisesCredentials) SensitiveStrings() []string {
	return []string{
		c.Username,
		c.Password,
	}
}
