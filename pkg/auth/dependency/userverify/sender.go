package userverify

import (
	"net/url"
	"path"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type CodeSender interface {
	Send(verifyCode VerifyCode, user model.User) error
}

type EmailCodeSender struct {
	AppName        string
	URLPrefix      *url.URL
	EmailConfig    config.EmailMessageConfiguration
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

	var textBody string
	if textBody, err = e.TemplateEngine.RenderTemplate(
		TemplateItemTypeUserVerificationEmailTXT,
		context,
		template.ResolveOptions{
			Key: verifyCode.LoginIDKey,
		},
	); err != nil {
		err = errors.Newf("failed to render user verification text email: %w", err)
		return
	}

	var htmlBody string
	if htmlBody, err = e.TemplateEngine.RenderTemplate(
		TemplateItemTypeUserVerificationEmailHTML,
		context,
		template.ResolveOptions{
			Key: verifyCode.LoginIDKey,
		},
	); err != nil {
		err = errors.Newf("failed to render user verification HTML email: %w", err)
		return
	}

	err = e.Sender.Send(mail.SendOptions{
		MessageConfig: e.EmailConfig,
		Recipient:     verifyCode.LoginID,
		TextBody:      textBody,
		HTMLBody:      htmlBody,
	})
	if err != nil {
		err = errors.Newf("failed to send user verification email: %w", err)
	}

	return
}

type SMSCodeSender struct {
	AppName        string
	URLPrefix      *url.URL
	SMSConfig      config.SMSMessageConfiguration
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
	if textBody, err = t.TemplateEngine.RenderTemplate(
		TemplateItemTypeUserVerificationSMSTXT,
		context,
		template.ResolveOptions{
			Key: verifyCode.LoginIDKey,
		},
	); err != nil {
		err = errors.Newf("failed to render user verification SMS message: %w", err)
		return
	}

	err = t.SMSClient.Send(sms.SendOptions{
		MessageConfig: t.SMSConfig,
		To:            verifyCode.LoginID,
		Body:          textBody,
	})
	if err != nil {
		err = errors.Newf("failed to send user verification SMS message: %w", err)
	}

	return
}

func prepareVerifyRequestContext(
	verifyCode VerifyCode,
	appName string,
	urlPrefix *url.URL,
	user model.User,
) map[string]interface{} {
	verifyLink := *urlPrefix
	verifyLink.Path = path.Join(verifyLink.Path, "_auth/verify_code_form")
	verifyLink.RawQuery = url.Values{"code": []string{verifyCode.Code}, "user_id": []string{user.ID}}.Encode()

	return map[string]interface{}{
		"appname":      appName,
		"login_id_key": verifyCode.LoginIDKey,
		"login_id":     verifyCode.LoginID,
		"user":         user,
		"user_id":      user.ID,
		"code":         verifyCode.Code,
		"link":         verifyLink.String(),
	}
}
