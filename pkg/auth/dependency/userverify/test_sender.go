package userverify

import (
	"net/url"

	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type TestCodeSenderFactory interface {
	NewTestCodeSender(
		provider config.UserVerificationProvider,
		keyConfig config.UserVerificationProviderConfiguration,
		loginIDKey string,
		templates map[string]string,
	) CodeSender
}

type defaultTestCodeSenderFactory struct {
	Config         config.TenantConfiguration
	URLPrefix      *url.URL
	TemplateEngine *template.Engine
}

func NewDefaultUserVerifyTestCodeSenderFactory(c config.TenantConfiguration, urlPrefix *url.URL, templateEngine *template.Engine) TestCodeSenderFactory {
	return &defaultTestCodeSenderFactory{
		Config:         c,
		URLPrefix:      urlPrefix,
		TemplateEngine: templateEngine,
	}
}

func (d *defaultTestCodeSenderFactory) NewTestCodeSender(
	provider config.UserVerificationProvider,
	keyConfig config.UserVerificationProviderConfiguration,
	loginIDKey string,
	templates map[string]string,
) (codeSender CodeSender) {
	loader := template.NewStringLoader()
	for templateType, templateBody := range templates {
		loader.StringMap[authTemplate.VerifyTemplateNameForKey(loginIDKey, templateType)] = templateBody
	}
	templateEngine := d.TemplateEngine
	templateEngine.PrependLoader(loader)

	switch provider {
	case config.UserVerificationProviderSMTP:
		codeSender = &EmailCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.URLPrefix,
			ProviderConfig: keyConfig,
			Sender:         mail.NewSender(d.Config.UserConfig.SMTP),
			TemplateEngine: templateEngine,
		}

	case config.UserVerificationProviderTwilio:
		codeSender = &SMSCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.URLPrefix,
			SMSClient:      sms.NewTwilioClient(d.Config.UserConfig.Twilio),
			TemplateEngine: templateEngine,
		}

	case config.UserVerificationProviderNexmo:
		codeSender = &SMSCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.URLPrefix,
			SMSClient:      sms.NewNexmoClient(d.Config.UserConfig.Nexmo),
			TemplateEngine: templateEngine,
		}
	}

	return
}
