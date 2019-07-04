package userverify

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type CodeSenderFactory interface {
	NewCodeSender(key string) CodeSender
}

type DefaultCodeSenderFactory struct {
	CodeSenderMap map[string]CodeSender
}

func NewDefaultUserVerifyCodeSenderFactory(c config.TenantConfiguration, templateEngine *template.Engine) CodeSenderFactory {
	userVerifyConfig := c.UserConfig.UserVerification
	f := DefaultCodeSenderFactory{
		CodeSenderMap: map[string]CodeSender{},
	}
	for key, config := range userVerifyConfig.LoginIDKeys {
		var codeSender CodeSender
		switch config.Provider {
		case "smtp":
			codeSender = &EmailCodeSender{
				AppName:        c.AppName,
				Config:         userVerifyConfig,
				Dialer:         mail.NewDialer(c.AppConfig.SMTP),
				TemplateEngine: templateEngine,
			}
		case "twilio":
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				Config:         userVerifyConfig,
				SMSClient:      sms.NewTwilioClient(c.AppConfig.Twilio),
				TemplateEngine: templateEngine,
			}
		case "nexmo":
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				Config:         userVerifyConfig,
				SMSClient:      sms.NewNexmoClient(c.AppConfig.Nexmo),
				TemplateEngine: templateEngine,
			}
		default:
			panic(errors.New("invalid user verify provider: " + config.Provider))
		}
		f.CodeSenderMap[key] = codeSender
	}

	return &f
}

func (d *DefaultCodeSenderFactory) NewCodeSender(key string) CodeSender {
	return d.CodeSenderMap[key]
}
