package userverify

import (
	"errors"
	"strconv"

	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type CodeSenderFactory interface {
	NewCodeSender(key string) CodeSender
}

type TestCodeSenderFactory interface {
	NewTestCodeSender(key string, providerSettings map[string]string, templates map[string]string) TestCodeSender
}

type DefaultCodeSenderFactory struct {
	CodeSenderMap map[string]CodeSender
}

func NewDefaultUserVerifyCodeSenderFactory(c config.TenantConfiguration, templateEngine *template.Engine) CodeSenderFactory {
	userVerifyConfig := c.UserVerify
	f := DefaultCodeSenderFactory{
		CodeSenderMap: map[string]CodeSender{},
	}
	for _, keyConfig := range userVerifyConfig.KeyConfigs {
		var codeSender CodeSender
		switch keyConfig.Provider {
		case "smtp":
			codeSender = &EmailCodeSender{
				AppName:        c.AppName,
				Config:         userVerifyConfig,
				Dialer:         mail.NewDialer(c.SMTP),
				TemplateEngine: templateEngine,
			}
		case "twilio":
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				Config:         userVerifyConfig,
				SMSClient:      sms.NewTwilioClient(c.Twilio),
				TemplateEngine: templateEngine,
			}
		case "nexmo":
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				Config:         userVerifyConfig,
				SMSClient:      sms.NewNexmoClient(c.Nexmo),
				TemplateEngine: templateEngine,
			}
		default:
			panic(errors.New("invalid user verify provider: " + keyConfig.Provider))
		}
		f.CodeSenderMap[keyConfig.Key] = codeSender
	}

	return &f
}

func (d *DefaultCodeSenderFactory) NewCodeSender(key string) CodeSender {
	return d.CodeSenderMap[key]
}

type DefaultTestCodeSenderFactory struct {
	SMTPTestCodeSender   *TestEmailCodeSender
	TwilioTestCodeSender *TestSMSCodeSender
	NexmoTestCodeSender  *TestSMSCodeSender

	// Base template engine, to be merged with user provided templates
	BaseTemplateEngine *template.Engine
}

func NewDefaultUserVerifyTestCodeSenderFactory(c config.TenantConfiguration, templateEngine *template.Engine) TestCodeSenderFactory {
	return &DefaultTestCodeSenderFactory{
		SMTPTestCodeSender: &TestEmailCodeSender{
			AppName:   c.AppName,
			URLPrefix: c.URLPrefix,
			Dialer:    mail.NewDialer(c.SMTP),
		},
		BaseTemplateEngine: templateEngine,
	}
}

func (d *DefaultTestCodeSenderFactory) NewTestCodeSender(key string, providerSettings map[string]string, templates map[string]string) TestCodeSender {
	// add template string loader, which stores template value provided by user
	loader := template.NewStringLoader()
	for templateType, template := range templates {
		loader.StringMap[authTemplate.VerifyTemplateNameForKey(key, templateType)] = template
	}
	templateEngine := d.BaseTemplateEngine
	templateEngine.PrependLoader(loader)

	providerName := providerSettings["name"]
	switch providerName {
	case "smtp":
		dialer := d.SMTPTestCodeSender.Dialer
		smtpConfig := smtpConfigFromProviderSettings(providerSettings)
		if smtpConfig != nil {
			dialer = mail.NewDialer(*smtpConfig)
		}
		return &TestEmailCodeSender{
			AppName:   d.SMTPTestCodeSender.AppName,
			URLPrefix: d.SMTPTestCodeSender.URLPrefix,
			Config: TestEmailCodeSenderConfig{
				Sender:      providerSettings["smtp_sender"],
				SenderName:  providerSettings["smtp_sender_name"],
				Subject:     providerSettings["smtp_subject"],
				ReplyTo:     providerSettings["smtp_reply_to"],
				ReplyToName: providerSettings["smtp_reply_to_name"],
			},
			Dialer:         dialer,
			TemplateEngine: templateEngine,
		}
	// TODO: twilio, nexmo
	default:
		return nil
	}
}

func smtpConfigFromProviderSettings(providerSettings map[string]string) *config.SMTPConfiguration {
	if providerSettings["smtp_host"] == "" {
		return nil
	}

	port, err := strconv.Atoi(providerSettings["smtp_port"])
	if err != nil {
		panic(errors.New("invalid smtp_port in provider settings"))
	}

	return &config.SMTPConfiguration{
		Host:     providerSettings["smtp_host"],
		Port:     port,
		Mode:     providerSettings["smtp_mode"],
		Login:    providerSettings["smtp_login"],
		Password: providerSettings["smtp_password"],
	}
}
