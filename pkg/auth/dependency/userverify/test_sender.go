package userverify

import (
	"errors"
	"strconv"

	"github.com/sirupsen/logrus"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type TestCodeSenderFactory interface {
	NewTestCodeSender(
		provider string,
		providerConfig map[string]string,
		keyConfig config.UserVerificationProviderConfiguration,
		loginIDKey string,
		templates map[string]string,
	) CodeSender
}

type defaultTestCodeSenderFactory struct {
	Config         config.TenantConfiguration
	TemplateEngine *template.Engine
	Logger         *logrus.Entry
}

func NewDefaultUserVerifyTestCodeSenderFactory(c config.TenantConfiguration, templateEngine *template.Engine, logger *logrus.Entry) TestCodeSenderFactory {
	return &defaultTestCodeSenderFactory{
		Config:         c,
		TemplateEngine: templateEngine,
		Logger:         logger,
	}
}

func (d *defaultTestCodeSenderFactory) NewTestCodeSender(
	provider string,
	providerConfig map[string]string,
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
	case "smtp":
		smtpConfig := d.Config.AppConfig.SMTP
		for key, value := range providerConfig {
			switch key {
			case "host":
				smtpConfig.Host = value
			case "port":
				port, err := strconv.Atoi(value)
				if err != nil {
					panic(errors.New("invalid smtp_port in provider settings"))
				}
				smtpConfig.Port = port
			case "mode":
				smtpConfig.Mode = value
			case "login":
				smtpConfig.Login = value
			case "password":
				smtpConfig.Password = value
			}
		}
		codeSender = &EmailCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.Config.UserConfig.UserVerification.URLPrefix,
			ProviderConfig: keyConfig,
			Dialer:         mail.NewDialer(smtpConfig),
			TemplateEngine: templateEngine,
		}

	case "twilio":
		twilioConfig := d.Config.AppConfig.Twilio
		for key, value := range providerConfig {
			switch key {
			case "account_sid":
				twilioConfig.AccountSID = value
			case "auth_token":
				twilioConfig.AuthToken = value
			case "from":
				twilioConfig.From = value
			}
		}
		codeSender = &SMSCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.Config.UserConfig.UserVerification.URLPrefix,
			SMSClient:      sms.NewTwilioClient(twilioConfig),
			TemplateEngine: templateEngine,
		}

	case "nexmo":
		nexmoConfig := d.Config.AppConfig.Nexmo
		for key, value := range providerConfig {
			switch key {
			case "api_key":
				nexmoConfig.APIKey = value
			case "secret":
				nexmoConfig.APISecret = value
			case "from":
				nexmoConfig.From = value
			}
		}
		codeSender = &SMSCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.Config.UserConfig.UserVerification.URLPrefix,
			SMSClient:      sms.NewNexmoClient(nexmoConfig),
			TemplateEngine: templateEngine,
		}

	case "debug":
		codeSender = &DebugCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.Config.UserConfig.UserVerification.URLPrefix,
			TemplateEngine: templateEngine,
			Logger:         d.Logger,
		}

	default:
		panic(errors.New("invalid user verify provider: " + provider))
	}

	return
}
