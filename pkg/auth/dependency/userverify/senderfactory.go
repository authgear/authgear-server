package userverify

import (
	"errors"

	"github.com/sirupsen/logrus"
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

func NewDefaultUserVerifyCodeSenderFactory(c config.TenantConfiguration, templateEngine *template.Engine, logger *logrus.Entry) CodeSenderFactory {
	userVerifyConfig := c.UserConfig.UserVerification
	f := defaultCodeSenderFactory{
		CodeSenderMap: map[string]CodeSender{},
	}

	for key, config := range userVerifyConfig.LoginIDKeys {
		var codeSender CodeSender
		switch config.Provider {
		case "smtp":
			codeSender = &EmailCodeSender{
				AppName:        c.AppName,
				URLPrefix:      userVerifyConfig.URLPrefix,
				ProviderConfig: config.ProviderConfig,
				Dialer:         mail.NewDialer(c.AppConfig.SMTP),
				TemplateEngine: templateEngine,
			}
		case "twilio":
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				URLPrefix:      userVerifyConfig.URLPrefix,
				SMSClient:      sms.NewTwilioClient(c.AppConfig.Twilio),
				TemplateEngine: templateEngine,
			}
		case "nexmo":
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				URLPrefix:      userVerifyConfig.URLPrefix,
				SMSClient:      sms.NewNexmoClient(c.AppConfig.Nexmo),
				TemplateEngine: templateEngine,
			}
		case "debug":
			codeSender = &DebugCodeSender{
				AppName:        c.AppName,
				URLPrefix:      userVerifyConfig.URLPrefix,
				TemplateEngine: templateEngine,
				Logger:         logger,
			}
		default:
			panic(errors.New("invalid user verify provider: " + config.Provider))
		}
		f.CodeSenderMap[key] = codeSender
	}

	return &f
}

func (d *defaultCodeSenderFactory) NewCodeSender(loginIDKey string) CodeSender {
	return d.CodeSenderMap[loginIDKey]
}
