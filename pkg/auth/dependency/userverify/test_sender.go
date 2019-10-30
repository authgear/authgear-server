package userverify

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type TestCodeSenderFactory interface {
	NewTestCodeSender(
		messageHeader config.MessageHeader,
		loginIDKey string,
		templates map[string]string,
	) CodeSender
}

type defaultTestCodeSenderFactory struct {
	Config         config.TenantConfiguration
	URLPrefix      *url.URL
	TemplateEngine *template.Engine
	SMSClient      sms.Client
	MailSender     mail.Sender
}

func NewDefaultUserVerifyTestCodeSenderFactory(
	c config.TenantConfiguration,
	urlPrefix *url.URL,
	templateEngine *template.Engine,
	mailSender mail.Sender,
	smsClient sms.Client,
) TestCodeSenderFactory {
	return &defaultTestCodeSenderFactory{
		Config:         c,
		URLPrefix:      urlPrefix,
		TemplateEngine: templateEngine,
		SMSClient:      smsClient,
		MailSender:     mailSender,
	}
}

func (d *defaultTestCodeSenderFactory) NewTestCodeSender(
	messageHeader config.MessageHeader,
	loginIDKey string,
	templates map[string]string,
) (codeSender CodeSender) {
	// TODO(template): Unbreak test code sender
	// for templateType, templateBody := range templates {
	// 	loader.StringMap[authTemplate.VerifyTemplateNameForKey(loginIDKey, templateType)] = templateBody
	// }
	templateEngine := d.TemplateEngine

	authLoginIDKey, ok := d.Config.UserConfig.Auth.GetLoginIDKey(loginIDKey)
	if !ok {
		panic("userverify: invalid login id key: " + loginIDKey)
	}
	keyType := authLoginIDKey.Type
	metadataKey, _ := keyType.MetadataKey()

	switch metadataKey {
	case metadata.Email:
		codeSender = &EmailCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.URLPrefix,
			MessageHeader:  messageHeader,
			Sender:         d.MailSender,
			TemplateEngine: templateEngine,
		}

	case metadata.Phone:
		codeSender = &SMSCodeSender{
			AppName:        d.Config.AppName,
			URLPrefix:      d.URLPrefix,
			SMSClient:      d.SMSClient,
			TemplateEngine: templateEngine,
		}
	default:
		panic("userverify: unknown metadata key: " + metadataKey)
	}

	return
}
