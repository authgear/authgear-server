package translation

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/template"
)

const TemplateItemTypeTranslationJSON string = "translation.json"

var TemplateTranslationJSON = template.Register(template.T{
	Type: TemplateItemTypeTranslationJSON,
})

type Service struct {
	Context           context.Context
	EnvironmentConfig *config.EnvironmentConfig
	TemplateEngine    *template.Engine

	translations *template.TranslationMap `wire:"-"`
}

func (s *Service) translationMap() (*template.TranslationMap, error) {
	if s.translations == nil {
		preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
		validatorOptions := []template.ValidatorOption{
			template.AllowRangeNode(true),
			template.AllowTemplateNode(true),
			template.AllowDeclaration(true),
			template.MaxDepth(15),
		}

		renderCtx := &template.RenderContext{
			PreferredLanguageTags: preferredLanguageTags,
			ValidatorOptions:      validatorOptions,
		}

		t, err := s.TemplateEngine.Translation(
			renderCtx,
			TemplateItemTypeTranslationJSON,
		)
		if err != nil {
			return nil, err
		}
		s.translations = t
	}
	return s.translations, nil
}

func (s *Service) renderTemplate(typ string, args interface{}) (string, error) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)

	renderCtx := &template.RenderContext{
		PreferredLanguageTags: preferredLanguageTags,
	}

	out, err := s.TemplateEngine.Render(renderCtx, typ, args)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (s *Service) AppMetadata() (*AppMetadata, error) {
	t, err := s.translationMap()
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"StaticAssetURLPrefix": s.EnvironmentConfig.StaticAssetURLPrefix,
	}

	appName, err := t.RenderText("app.app-name", args)
	if err != nil {
		return nil, err
	}

	logoURI, err := t.RenderText("app.logo-uri", args)
	if err != nil {
		return nil, err
	}

	return &AppMetadata{
		AppName: appName,
		LogoURI: logoURI,
	}, nil
}

func (s *Service) emailMessageHeader(name string, args interface{}) (sender, replyTo, subject string, err error) {
	t, err := s.translationMap()
	if err != nil {
		return
	}

	sender, err = t.RenderText(fmt.Sprintf("email.%s.sender", name), args)
	if template.IsNotFound(err) {
		sender, err = t.RenderText("email.default.sender", args)
	}
	if err != nil {
		return
	}

	replyTo, err = t.RenderText(fmt.Sprintf("email.%s.reply-to", name), args)
	if template.IsNotFound(err) {
		replyTo, err = t.RenderText("email.default.reply-to", args)
	}
	if err != nil {
		return
	}

	subject, err = t.RenderText(fmt.Sprintf("email.%s.subject", name), args)
	if template.IsNotFound(err) {
		subject, err = t.RenderText("email.default.subject", args)
	}
	if err != nil {
		return
	}

	return
}

func (s *Service) EmailMessageData(msg *MessageSpec, args interface{}) (*EmailMessageData, error) {
	sender, replyTo, subject, err := s.emailMessageHeader(msg.Name, args)
	if err != nil {
		return nil, err
	}

	textBody, err := s.renderTemplate(msg.TXTEmailType, args)
	if err != nil {
		return nil, err
	}

	htmlBody, err := s.renderTemplate(msg.HTMLEmailType, args)
	if err != nil {
		return nil, err
	}

	return &EmailMessageData{
		Sender:   sender,
		ReplyTo:  replyTo,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}, nil
}

func (s *Service) smsMessageHeader(name string, args interface{}) (sender string, err error) {
	t, err := s.translationMap()
	if err != nil {
		return
	}

	sender, err = t.RenderText(fmt.Sprintf("sms.%s.sender", name), args)
	if template.IsNotFound(err) {
		sender, err = t.RenderText("sms.default.sender", args)
	}
	if err != nil {
		return
	}

	return
}

func (s *Service) SMSMessageData(msg *MessageSpec, args interface{}) (*SMSMessageData, error) {
	sender, err := s.smsMessageHeader(msg.Name, args)
	if err != nil {
		return nil, err
	}

	body, err := s.renderTemplate(msg.SMSType, args)
	if err != nil {
		return nil, err
	}

	return &SMSMessageData{
		Sender: sender,
		Body:   body,
	}, nil
}
