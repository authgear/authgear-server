package mfa

import (
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type senderImpl struct {
	smsClient      sms.Client
	mailSender     mail.Sender
	templateEngine *template.Engine
}

func NewSender(smsClient sms.Client, mailSender mail.Sender, templateEngine *template.Engine) Sender {
	return &senderImpl{
		smsClient:      smsClient,
		mailSender:     mailSender,
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
	body, err := s.templateEngine.RenderTextTemplate(
		TemplateItemTypeMFAOOBCodeSMSTXT,
		context,
		template.RenderOptions{Required: true},
	)
	if err != nil {
		err = errors.Newf("failed to render MFA SMS message: %w", err)
		return err
	}

	err = s.smsClient.Send(phone, body)
	if err != nil {
		err = errors.Newf("failed to send MFA SMS message: %w", err)
	}
	return err
}

func (s *senderImpl) SendEmail(context map[string]interface{}, email string) error {
	textBody, err := s.templateEngine.RenderTextTemplate(
		TemplateItemTypeMFAOOBCodeEmailTXT,
		context,
		template.RenderOptions{Required: true},
	)
	if err != nil {
		err = errors.Newf("failed to render MFA text email: %w", err)
		return err
	}

	htmlBody, err := s.templateEngine.RenderHTMLTemplate(
		TemplateItemTypeMFAOOBCodeEmailHTML,
		context,
		template.RenderOptions{Required: false},
	)
	if err != nil {
		err = errors.Newf("failed to render MFA HTML email: %w", err)
		return err
	}

	err = s.mailSender.Send(mail.SendOptions{
		// TODO(mfa): configurable email headers
		Sender:    "no-reply@skygeario.com",
		Recipient: email,
		Subject:   "MFA Code",
		ReplyTo:   "no-reply@skygeario.com",
		TextBody:  textBody,
		HTMLBody:  htmlBody,
	})
	if err != nil {
		err = errors.Newf("failed to send MFA email: %w", err)
		return err
	}

	return nil
}
