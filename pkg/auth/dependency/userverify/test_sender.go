package userverify

import (
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
	TemplateEngine *template.Engine
}

func NewDefaultUserVerifyTestCodeSenderFactory(c config.TenantConfiguration, templateEngine *template.Engine) TestCodeSenderFactory {
	return &defaultTestCodeSenderFactory{
		Config:         c,
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
			URLPrefix:      d.Config.UserConfig.UserVerification.URLPrefix,
			ProviderConfig: keyConfig,
			Dialer:         mail.NewDialer(d.Config.AppConfig.SMTP),
			TemplateEngine: templateEngine,
		}

	case config.UserVerificationProviderTwilio:
		codeSender = &SMSCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.Config.UserConfig.UserVerification.URLPrefix,
			SMSClient:      sms.NewTwilioClient(d.Config.AppConfig.Twilio),
			TemplateEngine: templateEngine,
		}

	case config.UserVerificationProviderNexmo:
		codeSender = &SMSCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.Config.UserConfig.UserVerification.URLPrefix,
			SMSClient:      sms.NewNexmoClient(d.Config.AppConfig.Nexmo),
			TemplateEngine: templateEngine,
		}
	}

	return
}
