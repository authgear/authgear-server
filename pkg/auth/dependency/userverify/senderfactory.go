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
	userVerifyConfig := c.UserConfig.UserVerification
	f := DefaultCodeSenderFactory{
		CodeSenderMap: map[string]CodeSender{},
	}
	for _, keyConfig := range userVerifyConfig.Keys {
		var codeSender CodeSender
		switch keyConfig.Provider {
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
			URLPrefix: c.UserConfig.URLPrefix,
			Dialer:    mail.NewDialer(c.AppConfig.SMTP),
		},
		TwilioTestCodeSender: &TestSMSCodeSender{
			AppName:   c.AppName,
			URLPrefix: c.UserConfig.URLPrefix,
			SMSClient: sms.NewTwilioClient(c.AppConfig.Twilio),
		},
		NexmoTestCodeSender: &TestSMSCodeSender{
			AppName:   c.AppName,
			URLPrefix: c.UserConfig.URLPrefix,
			SMSClient: sms.NewNexmoClient(c.AppConfig.Nexmo),
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
	case "twilio":
		smsClient := d.TwilioTestCodeSender.SMSClient
		smsClientConfig := twilioSmsClientConfigFromProviderSettings(providerSettings)
		if smsClientConfig != nil {
			smsClient = sms.NewTwilioClient(*smsClientConfig)
		}
		return &TestSMSCodeSender{
			AppName:        d.TwilioTestCodeSender.AppName,
			URLPrefix:      d.TwilioTestCodeSender.URLPrefix,
			SMSClient:      smsClient,
			TemplateEngine: templateEngine,
		}
	case "nexmo":
		smsClient := d.NexmoTestCodeSender.SMSClient
		smsClientConfig := nexmoSmsClientConfigFromProviderSettings(providerSettings)
		if smsClientConfig != nil {
			smsClient = sms.NewNexmoClient(*smsClientConfig)
		}
		return &TestSMSCodeSender{
			AppName:        d.NexmoTestCodeSender.AppName,
			URLPrefix:      d.NexmoTestCodeSender.URLPrefix,
			SMSClient:      smsClient,
			TemplateEngine: templateEngine,
		}
	default:
		return nil
	}
}

func smtpConfigFromProviderSettings(providerSettings map[string]string) *config.NewSMTPConfiguration {
	if providerSettings["smtp_host"] == "" {
		return nil
	}

	port, err := strconv.Atoi(providerSettings["smtp_port"])
	if err != nil {
		panic(errors.New("invalid smtp_port in provider settings"))
	}

	return &config.NewSMTPConfiguration{
		Host:     providerSettings["smtp_host"],
		Port:     port,
		Mode:     providerSettings["smtp_mode"],
		Login:    providerSettings["smtp_login"],
		Password: providerSettings["smtp_password"],
	}
}

func twilioSmsClientConfigFromProviderSettings(providerSettings map[string]string) *config.NewTwilioConfiguration {
	if providerSettings["twilio_account_sid"] == "" {
		return nil
	}

	return &config.NewTwilioConfiguration{
		AccountSID: providerSettings["twilio_account_sid"],
		AuthToken:  providerSettings["twilio_auth_token"],
		From:       providerSettings["twilio_from"],
	}
}

func nexmoSmsClientConfigFromProviderSettings(providerSettings map[string]string) *config.NewNexmoConfiguration {
	if providerSettings["nexmo_api_key"] == "" {
		return nil
	}

	return &config.NewNexmoConfiguration{
		APIKey:    providerSettings["nexmo_api_key"],
		APISecret: providerSettings["nexmo_api_secret"],
		From:      providerSettings["nexmo_from"],
	}
}
