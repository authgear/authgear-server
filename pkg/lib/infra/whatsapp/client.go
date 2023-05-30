package whatsapp

import (
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/sirupsen/logrus"
)

type ClientLogger struct{ *log.Logger }

func NewClientLogger(lf *log.Factory) ClientLogger { return ClientLogger{lf.New("whatsapp-client")} }

type Client struct {
	Logger                     ClientLogger
	DevMode                    config.DevMode
	TestModeWhatsappSuppressed config.TestModeWhatsappSuppressed
	Config                     *config.WhatsappConfig
	OnPremisesClient           *OnPremisesClient
	TokenStore                 *TokenStore
}

func (c *Client) logMessage(
	to string,
	templateName string,
	templateLanguage string,
	templateComponents []TemplateComponent,
	namespace string) *logrus.Entry {
	data, _ := json.MarshalIndent(templateComponents, "", "  ")

	return c.Logger.
		WithField("recipient", to).
		WithField("template_name", templateName).
		WithField("language", templateLanguage).
		WithField("components", string(data)).
		WithField("namespace", namespace)
}

func (c *Client) SendTemplate(
	to string,
	templateName string,
	templateLanguage string,
	templateComponents []TemplateComponent,
	namespace string,
) error {
	if c.TestModeWhatsappSuppressed {
		c.logMessage(to, templateName, templateLanguage, templateComponents, namespace).
			Warn("sending whatsapp is suppressed in test mode")
		return nil
	}

	if c.DevMode {
		c.logMessage(to, templateName, templateLanguage, templateComponents, namespace).
			Warn("skip sending whatsapp in development mode")
		return nil
	}

	switch c.Config.APIType {
	case config.WhatsappAPITypeOnPremises:
		if c.OnPremisesClient == nil {
			return ErrNoAvailableClient
		}
		return c.OnPremisesClient.SendTemplate(to, templateName, templateLanguage, templateComponents, namespace)
	default:
		return fmt.Errorf("whatsapp: unknown api type")
	}
}
