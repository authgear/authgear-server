package userverify

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/sms"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type CodeSender interface {
	Send(verifyCode VerifyCode, user model.User) error
}

type EmailCodeSender struct {
	AppName        string
	URLPrefix      string
	ProviderConfig config.UserVerificationProviderConfiguration
	Sender         mail.Sender
	TemplateEngine *template.Engine
}

func (e *EmailCodeSender) Send(verifyCode VerifyCode, user model.User) (err error) {
	context := prepareVerifyRequestContext(
		verifyCode,
		e.AppName,
		e.URLPrefix,
		user,
	)

	providerConfig := e.ProviderConfig

	var textBody string
	if textBody, err = e.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyTextTemplateNameForKey(verifyCode.LoginIDKey),
		context,
		template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifyEmailText},
	); err != nil {
		return
	}

	var htmlBody string
	if htmlBody, err = e.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyHTMLTemplateNameForKey(verifyCode.LoginIDKey),
		context,
		template.ParseOption{Required: false, FallbackTemplateName: authTemplate.TemplateNameVerifyEmailHTML},
	); err != nil {
		return
	}

	err = e.Sender.Send(mail.SendOptions{
		Sender:    providerConfig.Sender,
		Recipient: verifyCode.LoginID,
		Subject:   providerConfig.Subject,
		ReplyTo:   providerConfig.ReplyTo,
		TextBody:  textBody,
		HTMLBody:  htmlBody,
	})

	return
}

type SMSCodeSender struct {
	AppName        string
	URLPrefix      string
	SMSClient      sms.Client
	TemplateEngine *template.Engine
}

func (t *SMSCodeSender) Send(verifyCode VerifyCode, user model.User) (err error) {
	context := prepareVerifyRequestContext(
		verifyCode,
		t.AppName,
		t.URLPrefix,
		user,
	)

	var textBody string
	if textBody, err = t.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyTextTemplateNameForKey(verifyCode.LoginIDKey),
		context,
		template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifySMSText},
	); err != nil {
		return
	}

	err = t.SMSClient.Send(verifyCode.LoginID, textBody)
	return
}

func prepareVerifyRequestContext(
	verifyCode VerifyCode,
	appName string,
	urlPrefix string,
	user model.User,
) map[string]interface{} {
	return map[string]interface{}{
		"appname":      appName,
		"login_id_key": verifyCode.LoginIDKey,
		"login_id":     verifyCode.LoginID,
		"user":         user,
		"code":         verifyCode.Code,
		"link": fmt.Sprintf(
			"%s/verify_code_form?code=%s&user_id=%s",
			urlPrefix,
			verifyCode.Code,
			user.ID,
		),
	}
}
