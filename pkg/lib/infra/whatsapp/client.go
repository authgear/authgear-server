package whatsapp

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Client struct {
	Config           *config.WhatsappConfig
	OnPremisesClient *OnPremisesClient
	TokenStore       *TokenStore
}

func (c *Client) SendTemplate(
	to string,
	templateName string,
	templateLanguage string,
	templateComponents []TemplateComponent,
) error {
	switch c.Config.APIType {
	case config.WhatsappAPITypeOnPremises:
		if c.OnPremisesClient == nil {
			return ErrNoAvailableClient
		}
		return c.OnPremisesClient.SendTemplate(to, templateName, templateLanguage, templateComponents)
	default:
		return fmt.Errorf("whatsapp: unknown api type")
	}
}
