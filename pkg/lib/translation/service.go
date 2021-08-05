package translation

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type StaticAssetResolver interface {
	StaticAssetURL(id string) (url string, err error)
}

type Service struct {
	Context        context.Context
	TemplateEngine *template.Engine
	StaticAssets   StaticAssetResolver

	translations *template.TranslationMap `wire:"-"`
}

func (s *Service) translationMap() (*template.TranslationMap, error) {
	if s.translations == nil {
		preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
		t, err := s.TemplateEngine.Translation(preferredLanguageTags)
		if err != nil {
			return nil, err
		}
		s.translations = t
	}
	return s.translations, nil
}

func (s *Service) appMetadata(data map[string]interface{}) error {
	t, err := s.translationMap()
	if err != nil {
		return err
	}

	// TODO(l10n): investigate on how to allow referencing other translation natively.
	appName, err := t.RenderText("app.name", nil)
	if err != nil {
		return err
	}

	data["AppName"] = appName

	return nil
}

func (s *Service) renderTemplate(tpl template.Resource, args interface{}) (string, error) {
	data := make(map[string]interface{})
	template.Embed(data, args)
	data["StaticAssetURL"] = s.StaticAssets.StaticAssetURL

	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)
	out, err := s.TemplateEngine.Render(tpl, preferredLanguageTags, data)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (s *Service) emailMessageHeader(name string, args interface{}) (sender, replyTo, subject string, err error) {
	t, err := s.translationMap()
	if err != nil {
		return
	}

	data := make(map[string]interface{})
	template.Embed(data, args)
	err = s.appMetadata(data)
	if err != nil {
		return
	}

	sender, err = t.RenderText(fmt.Sprintf("email.%s.sender", name), data)
	if errors.Is(err, template.ErrNotFound) {
		sender, err = t.RenderText("email.default.sender", data)
	}
	if err != nil {
		return
	}

	replyTo, err = t.RenderText(fmt.Sprintf("email.%s.reply-to", name), data)
	if errors.Is(err, template.ErrNotFound) {
		replyTo, err = t.RenderText("email.default.reply-to", data)
	}
	if err != nil {
		return
	}

	subject, err = t.RenderText(fmt.Sprintf("email.%s.subject", name), data)
	if errors.Is(err, template.ErrNotFound) {
		subject, err = t.RenderText("email.default.subject", data)
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

	textBody, err := s.renderTemplate(msg.TXTEmailTemplate, args)
	if err != nil {
		return nil, err
	}

	htmlBody, err := s.renderTemplate(msg.HTMLEmailTemplate, args)
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

	data := make(map[string]interface{})
	template.Embed(data, args)
	err = s.appMetadata(data)
	if err != nil {
		return
	}

	sender, err = t.RenderText(fmt.Sprintf("sms.%s.sender", name), data)
	if errors.Is(err, template.ErrNotFound) {
		sender, err = t.RenderText("sms.default.sender", data)
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

	body, err := s.renderTemplate(msg.SMSTemplate, args)
	if err != nil {
		return nil, err
	}

	return &SMSMessageData{
		Sender: sender,
		Body:   body,
	}, nil
}

func (s *Service) HasKey(key string) (bool, error) {
	t, err := s.translationMap()
	if err != nil {
		return false, err
	}
	return t.HasKey(key), nil
}

func (s *Service) RenderText(key string, args interface{}) (string, error) {
	t, err := s.translationMap()
	if err != nil {
		return "", err
	}
	return t.RenderText(key, args)
}
