package userverify

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
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
	var config config.UserVerificationKeyConfiguration
	var ok bool
	if config, ok = e.Config.LoginIDKeys[verifyCode.LoginIDKey]; !ok {
		return errors.New("provider for " + verifyCode.LoginIDKey + " not found")
	}

	context := prepareVerifyRequestContext(
		verifyCode,
		e.AppName,
		e.Config,
		user,
	)

	providerConfig := config.ProviderConfig

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

	sendReq := mail.SendRequest{
		Dialer:      e.Dialer,
		Sender:      providerConfig.Sender,
		SenderName:  providerConfig.SenderName,
		Recipient:   verifyCode.LoginID,
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
		authTemplate.VerifyTextTemplateNameForKey(verifyCode.LoginIDKey),
		context,
		template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifySMSText},
	); err != nil {
		return
	}

	err = t.SMSClient.Send(verifyCode.LoginID, textBody)
	return
}

type DebugCodeSender struct {
	AppName        string
	Config         config.UserVerificationConfiguration
	TemplateEngine *template.Engine
	Logger         *logrus.Entry
}

func (t *DebugCodeSender) Send(verifyCode VerifyCode, user response.User) (err error) {
	context := prepareVerifyRequestContext(
		verifyCode,
		t.AppName,
		t.Config,
		user,
	)

	var textBody string
	if textBody, err = t.TemplateEngine.ParseTextTemplate(
		authTemplate.VerifyTextTemplateNameForKey(verifyCode.LoginIDKey),
		context,
		template.ParseOption{Required: true, FallbackTemplateName: authTemplate.TemplateNameVerifyEmailText},
	); err != nil {
		return
	}

	t.Logger.WithFields(logrus.Fields{
		"body": textBody,
	}).Info("Send verification code")

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
		"login_id_key": verifyCode.LoginIDKey,
		"login_id":     verifyCode.LoginID,
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
