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
	Send(code string, userProfile userprofile.UserProfile) error
}

type EmailCodeSender struct {
	Key     string
	AppName string
	Config  config.UserVerifyConfiguration
	Dialer  *gomail.Dialer
	CodeGenerator
}

func (e *EmailCodeSender) Send(code string, userProfile userprofile.UserProfile) (err error) {
	var recordValue string
	var ok bool
	if recordValue, ok = userProfile.Data[e.Key].(string); !ok {
		return errors.New(e.Key + " is invalid in user data")
	}

	var keyConfig config.UserVerifyKeyConfiguration
	if keyConfig, ok = e.Config.ConfigForKey(e.Key); !ok {
		return errors.New("provider for " + e.Key + " not found")
	}

	context := prepareVerifyRequestContext(
		e.Key,
		recordValue,
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
		Recipient:   recordValue,
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
	Key       string
	AppName   string
	Config    config.UserVerifyConfiguration
	SMSClient sms.Client
	CodeGenerator
}

func (t *SMSCodeSender) Send(code string, userProfile userprofile.UserProfile) (err error) {
	var recordValue string
	var ok bool
	if recordValue, ok = userProfile.Data[t.Key].(string); !ok {
		return errors.New(t.Key + " is invalid in user data")
	}

	var keyConfig config.UserVerifyKeyConfiguration
	if keyConfig, ok = t.Config.ConfigForKey(t.Key); !ok {
		return errors.New("provider for " + t.Key + " not found")
	}

	context := prepareVerifyRequestContext(
		t.Key,
		recordValue,
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

	err = t.SMSClient.Send(recordValue, textBody)
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
