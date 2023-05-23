package config

var _ = SecretConfigSchema.Add("WhatsappOnPremisesCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"username": { "type": "string", "minLength": 1 },
		"password": { "type": "string", "minLength": 1 },
		"namespace": { "type": "string", "minLength": 1 }
	},
	"required": ["username", "password", "namespace"]
}
`)

type WhatsappOnPremisesCredentials struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Namespace string `json:"namespace"`
}

func (c *WhatsappOnPremisesCredentials) SensitiveStrings() []string {
	return []string{
		c.Password,
	}
}
