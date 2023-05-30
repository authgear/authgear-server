package config

var _ = SecretConfigSchema.Add("WhatsappOnPremisesCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"api_endpoint": { "type": "string", "minLength": 1 },
		"username": { "type": "string", "minLength": 1 },
		"password": { "type": "string", "minLength": 1 },
		"namespace": { "type": "string", "minLength": 1 }
	},
	"required": ["api_endpoint", "username", "password", "namespace"]
}
`)

type WhatsappOnPremisesCredentials struct {
	APIEndpoint string `json:"api_endpoint"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Namespace   string `json:"namespace"`
}

func (c *WhatsappOnPremisesCredentials) SensitiveStrings() []string {
	return []string{
		c.Password,
	}
}
