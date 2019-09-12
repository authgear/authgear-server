package mfa

import (
	"github.com/go-gomail/gomail"

	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type senderImpl struct {
	smsClient      sms.Client
	mailDialer     *gomail.Dialer
	templateEngine *template.Engine
}

func NewSender(smsClient sms.Client, mailDialer *gomail.Dialer, templateEngine *template.Engine) Sender {
	return &senderImpl{
		smsClient:      smsClient,
		mailDialer:     mailDialer,
		templateEngine: templateEngine,
	}
}

func (s *senderImpl) Send(code string, phone string, email string) error {
	context := map[string]interface{}{
		"code": code,
	}
	if phone != "" {
		return s.SendSMS(context, phone)
	}
	if email != "" {
		return s.SendEmail(context, email)
	}
	return nil
}

func (s *senderImpl) SendSMS(context map[string]interface{}, phone string) error {
	body, err := s.templateEngine.ParseTextTemplate(
		authTemplate.TemplateNameMFAOOBCodeSMSText,
		context,
		template.ParseOption{Required: true},
	)
	if err != nil {
		return err
	}
	return s.smsClient.Send(phone, body)
}

func (s *senderImpl) SendEmail(context map[string]interface{}, email string) error {
	textBody, err := s.templateEngine.ParseTextTemplate(
		authTemplate.TemplateNameMFAOOBCodeEmailText,
		context,
		template.ParseOption{Required: true},
	)
	if err != nil {
		return err
	}

	htmlBody, err := s.templateEngine.ParseHTMLTemplate(
		authTemplate.TemplateNameMFAOOBCodeEmailHTML,
		context,
		template.ParseOption{Required: false},
	)
	if err != nil {
		return err
	}

	sendRequest := mail.SendRequest{
		Dialer: s.mailDialer,
		// TODO(mfa): configurable email headers
		Sender:    "no-reply@skygeario.com",
		Recipient: email,
		Subject:   "MFA Code",
		ReplyTo:   "no-reply@skygeario.com",
		TextBody:  textBody,
		HTMLBody:  htmlBody,
	}
	err = sendRequest.Execute()
	if err != nil {
		return err
	}

	return nil
}
