package userverify

import (
	"errors"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
)

type CodeSenderFactory interface {
	NewCodeSender(key string) CodeSender
}

type DefaultCodeSenderFactory struct {
	CodeSenderMap map[string]CodeSender
}

func NewDefaultUserVerifyCodeSenderFactory(c config.TenantConfiguration) CodeSenderFactory {
	userVerifyConfig := c.UserVerify
	f := DefaultCodeSenderFactory{
		CodeSenderMap: map[string]CodeSender{},
	}
	for _, keyConfig := range userVerifyConfig.KeyConfigs {
		var codeSender CodeSender
		switch keyConfig.Provider {
		case "smtp":
			codeSender = &EmailCodeSender{
				AppName: c.AppName,
				// Key:           keyConfig.Key,
				Config:        userVerifyConfig,
				Dialer:        mail.NewDialer(c.SMTP),
				CodeGenerator: NewCodeGenerator(keyConfig.CodeFormat),
			}
		case "twilio":
			codeSender = &SMSCodeSender{
				AppName: c.AppName,
				// Key:           keyConfig.Key,
				Config:        userVerifyConfig,
				SMSClient:     sms.NewTwilioClient(c.Twilio),
				CodeGenerator: NewCodeGenerator(keyConfig.CodeFormat),
			}
		case "nexmo":
			codeSender = &SMSCodeSender{
				AppName: c.AppName,
				// Key:           keyConfig.Key,
				Config:        userVerifyConfig,
				SMSClient:     sms.NewNexmoClient(c.Nexmo),
				CodeGenerator: NewCodeGenerator(keyConfig.CodeFormat),
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
