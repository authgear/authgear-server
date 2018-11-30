package userverify

import (
	"errors"
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/sms"

	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type CodeSender interface {
	CodeGenerator
	Send(code string, key string, value string, userProfile userprofile.UserProfile) error
}

type EmailCodeSender struct {
	AppName string
	Config  config.UserVerifyConfiguration
	Dialer  *gomail.Dialer
	CodeGenerator
}

func (e *EmailCodeSender) Send(code string, key string, value string, userProfile userprofile.UserProfile) (err error) {
	var keyConfig config.UserVerifyKeyConfiguration
	var ok bool
	if keyConfig, ok = e.Config.ConfigForKey(key); !ok {
		return errors.New("provider for " + key + " not found")
	}

	context := prepareVerifyRequestContext(
		key,
		value,
		e.AppName,
		e.Config,
		code,
		userProfile,
	)

	providerConfig := keyConfig.ProviderConfig

	var textBody string
	if textBody, err = template.ParseTextTemplateFromURL(providerConfig.TextURL, context); err != nil {
		return
	}

	var htmlBody string
	if providerConfig.HTMLURL != "" {
		if htmlBody, err = template.ParseHTMLTemplateFromURL(providerConfig.HTMLURL, context); err != nil {
			return
		}
	}

	sendReq := mail.SendRequest{
		Dialer:      e.Dialer,
		Sender:      providerConfig.Sender,
		SenderName:  providerConfig.SenderName,
		Recipient:   value,
		Subject:     providerConfig.Subject,
		ReplyTo:     providerConfig.ReplyTo,
		ReplyToName: providerConfig.ReplyToName,
		TextBody:    textBody,
		HTMLBody:    htmlBody,
	}

	err = sendReq.Execute()
	return
}

type SMSCodeSender struct {
	AppName   string
	Config    config.UserVerifyConfiguration
	SMSClient sms.Client
	CodeGenerator
}

func (t *SMSCodeSender) Send(code string, key string, value string, userProfile userprofile.UserProfile) (err error) {
	var keyConfig config.UserVerifyKeyConfiguration
	var ok bool
	if keyConfig, ok = t.Config.ConfigForKey(key); !ok {
		return errors.New("provider for " + key + " not found")
	}

	context := prepareVerifyRequestContext(
		key,
		value,
		t.AppName,
		t.Config,
		code,
		userProfile,
	)

	providerConfig := keyConfig.ProviderConfig

	var textBody string
	if textBody, err = template.ParseTextTemplateFromURL(providerConfig.TextURL, context); err != nil {
		return
	}

	err = t.SMSClient.Send(value, textBody)
	return
}

func prepareVerifyRequestContext(
	key string,
	value string,
	appName string,
	config config.UserVerifyConfiguration,
	code string,
	userProfile userprofile.UserProfile,
) map[string]interface{} {
	return map[string]interface{}{
		"appname":      appName,
		"record_key":   key,
		"record_value": value,
		"user_id":      userProfile.RecordID,
		"user":         userProfile.ToMap(),
		"code":         code,
		"link": fmt.Sprintf(
			"%s/auth/verify-code/form?code=%s&user_id=%s",
			config.URLPrefix,
			code,
			userProfile.RecordID,
		),
	}
}
