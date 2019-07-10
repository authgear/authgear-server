package userverify

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type CodeSenderFactory interface {
	NewCodeSender(loginIDKey string) CodeSender
}

type defaultCodeSenderFactory struct {
	CodeSenderMap map[string]CodeSender
}

func NewDefaultUserVerifyCodeSenderFactory(c config.TenantConfiguration, templateEngine *template.Engine) CodeSenderFactory {
	userVerifyConfig := c.UserConfig.UserVerification
	f := defaultCodeSenderFactory{
		CodeSenderMap: map[string]CodeSender{},
	}

	for key, verifyConfig := range userVerifyConfig.LoginIDKeys {
		var codeSender CodeSender
		switch verifyConfig.Provider {
		case config.UserVerificationProviderSMTP:
			codeSender = &EmailCodeSender{
				AppName:        c.AppName,
				URLPrefix:      userVerifyConfig.URLPrefix,
				ProviderConfig: verifyConfig.ProviderConfig,
				Dialer:         mail.NewDialer(c.AppConfig.SMTP),
				TemplateEngine: templateEngine,
			}
		case config.UserVerificationProviderTwilio:
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				URLPrefix:      userVerifyConfig.URLPrefix,
				SMSClient:      sms.NewTwilioClient(c.AppConfig.Twilio),
				TemplateEngine: templateEngine,
			}
		case config.UserVerificationProviderNexmo:
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				URLPrefix:      userVerifyConfig.URLPrefix,
				SMSClient:      sms.NewNexmoClient(c.AppConfig.Nexmo),
				TemplateEngine: templateEngine,
			}
		}
		f.CodeSenderMap[key] = codeSender
	}

	return &f
}

func (d *defaultCodeSenderFactory) NewCodeSender(loginIDKey string) CodeSender {
	return d.CodeSenderMap[loginIDKey]
}
