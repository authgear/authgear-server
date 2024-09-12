package translation

import (
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/template"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package translation_test

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
	preferredLanguageTags := intl.GetPreferredLanguageTags(s.Context)

	return s.renderTemplateInLanguage(preferredLanguageTags, tpl, args)
}

func (s *Service) renderTemplateInLanguage(preferredLanguages []string, tpl template.Resource, args interface{}) (string, error) {
	data := make(map[string]interface{})
	template.Embed(data, args)
	data["StaticAssetURL"] = s.StaticAssets.StaticAssetURL

	out, err := s.TemplateEngine.Render(tpl, preferredLanguages, data)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (s *Service) GetSenderForTestEmail() (sender string, err error) {
	t, err := s.translationMap()
	if err != nil {
		return
	}

	sender, err = t.RenderText("email.default.sender", nil)
	if err != nil {
		return
	}

	return
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
	uiParam := uiparam.GetUIParam(s.Context)

	// Ensure these data are safe to put at query
	textData := map[string]interface{}{
		"ClientID":  htmltemplate.URLQueryEscaper(uiParam.ClientID),
		"State":     htmltemplate.URLQueryEscaper(uiParam.State),
		"XState":    htmltemplate.URLQueryEscaper(uiParam.XState),
		"UILocales": htmltemplate.URLQueryEscaper(uiParam.UILocales),
	}

	// html template will handle the escape
	htmlData := map[string]interface{}{
		"ClientID":  uiParam.ClientID,
		"State":     uiParam.State,
		"XState":    uiParam.XState,
		"UILocales": uiParam.UILocales,
	}
	template.Embed(htmlData, args)
	template.Embed(textData, args)

	sender, replyTo, subject, err := s.emailMessageHeader(msg.Name, htmlData)
	if err != nil {
		return nil, err
	}

	textBody, err := s.renderTemplate(msg.TXTEmailTemplate, textData)
	if err != nil {
		return nil, err
	}

	htmlBody, err := s.renderTemplate(msg.HTMLEmailTemplate, htmlData)
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
	uiParam := uiparam.GetUIParam(s.Context)
	data := map[string]interface{}{
		"ClientID":  htmltemplate.URLQueryEscaper(uiParam.ClientID),
		"State":     htmltemplate.URLQueryEscaper(uiParam.State),
		"XState":    htmltemplate.URLQueryEscaper(uiParam.XState),
		"UILocales": htmltemplate.URLQueryEscaper(uiParam.UILocales),
	}
	template.Embed(data, args)

	sender, err := s.smsMessageHeader(msg.Name, data)
	if err != nil {
		return nil, err
	}

	body, err := s.renderTemplate(msg.SMSTemplate, data)
	if err != nil {
		return nil, err
	}

	return &SMSMessageData{
		Sender: sender,
		Body:   body,
	}, nil
}

func (s *Service) WhatsappMessageData(language string, msg *MessageSpec, args interface{}) (*WhatsappMessageData, error) {
	data := map[string]interface{}{}
	template.Embed(data, args)

	body, err := s.renderTemplateInLanguage([]string{language}, msg.WhatsappTemplate, data)
	if err != nil {
		return nil, err
	}

	return &WhatsappMessageData{
		Body: body,
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
