package userverify

import (
	"errors"
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/sms"

	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/response"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type CodeSender interface {
	Send(verifyCode VerifyCode, user response.User) error
}

type EmailCodeSender struct {
	AppName        string
	Config         config.UserVerificationConfiguration
	Dialer         *gomail.Dialer
	TemplateEngine *template.Engine
}

func (e *EmailCodeSender) Send(verifyCode VerifyCode, user response.User) (err error) {
	var keyConfig config.UserVerificationKeyConfiguration
	var ok bool
	if keyConfig, ok = e.Config.ConfigForKey(verifyCode.RecordKey); !ok {
		return errors.New("provider for " + verifyCode.RecordKey + " not found")
	}

	context := prepareVerifyRequestContext(
		verifyCode,
		e.AppName,
		e.Config,
		user,
	)

	providerConfig := keyConfig.ProviderConfig

	var textBody string
	if textBody, err = e.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyTextTemplateNameForKey(verifyCode.RecordKey),
		context,
		template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifyEmailText},
	); err != nil {
		return
	}

	var htmlBody string
	if htmlBody, err = e.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyHTMLTemplateNameForKey(verifyCode.RecordKey),
		context,
		template.ParseOption{Required: false, FallbackTemplateName: authTemplate.TemplateNameVerifyEmailHTML},
	); err != nil {
		return
	}

	sendReq := mail.SendRequest{
		Dialer:      e.Dialer,
		Sender:      providerConfig.Sender,
		SenderName:  providerConfig.SenderName,
		Recipient:   verifyCode.RecordValue,
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
	AppName        string
	Config         config.UserVerificationConfiguration
	SMSClient      sms.Client
	TemplateEngine *template.Engine
}

func (t *SMSCodeSender) Send(verifyCode VerifyCode, user response.User) (err error) {
	context := prepareVerifyRequestContext(
		verifyCode,
		t.AppName,
		t.Config,
		user,
	)

	var textBody string
	if textBody, err = t.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyTextTemplateNameForKey(verifyCode.RecordKey),
		context,
		template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifySMSText},
	); err != nil {
		return
	}

	err = t.SMSClient.Send(verifyCode.RecordValue, textBody)
	return
}

func prepareVerifyRequestContext(
	verifyCode VerifyCode,
	appName string,
	config config.UserVerificationConfiguration,
	user response.User,
) map[string]interface{} {
	return map[string]interface{}{
		"appname":      appName,
		"record_key":   verifyCode.RecordKey,
		"record_value": verifyCode.RecordValue,
		"user_id":      user.UserID,
		"user":         user,
		"code":         verifyCode.Code,
		"link": fmt.Sprintf(
			"%s/verify_code_form?code=%s&user_id=%s",
			config.URLPrefix,
			verifyCode.Code,
			user.UserID,
		),
	}
}
