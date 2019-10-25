package userverify

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
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

func NewDefaultUserVerifyCodeSenderFactory(
	c config.TenantConfiguration,
	urlPrefix *url.URL,
	templateEngine *template.Engine,
	mailSender mail.Sender,
	smsClient sms.Client,
) CodeSenderFactory {
	userVerifyConfig := c.UserConfig.UserVerification
	f := defaultCodeSenderFactory{
		CodeSenderMap: map[string]CodeSender{},
	}

	for key, verifyConfig := range userVerifyConfig.LoginIDKeys {
		var codeSender CodeSender
		keyType := c.UserConfig.Auth.LoginIDKeys[key].Type
		metadataKey, _ := keyType.MetadataKey()
		switch metadataKey {
		case metadata.Email:
			codeSender = &EmailCodeSender{
				AppName:        c.AppName,
				URLPrefix:      urlPrefix,
				ProviderConfig: verifyConfig.ProviderConfig,
				Sender:         mailSender,
				TemplateEngine: templateEngine,
			}
		case metadata.Phone:
			codeSender = &SMSCodeSender{
				AppName:        c.AppName,
				URLPrefix:      urlPrefix,
				SMSClient:      smsClient,
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
