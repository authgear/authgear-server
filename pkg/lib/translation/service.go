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
	StaticAssetURL(ctx context.Context, id string) (url string, err error)
}

type Service struct {
	TemplateEngine *template.Engine
	StaticAssets   StaticAssetResolver

	translations *template.TranslationMap `wire:"-"`
}

func (s *Service) translationMap(ctx context.Context) (*template.TranslationMap, error) {
	if s.translations == nil {
		preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)
		t, err := s.TemplateEngine.Translation(preferredLanguageTags)
		if err != nil {
			return nil, err
		}
		s.translations = t
	}
	return s.translations, nil
}

func (s *Service) renderTemplate(ctx context.Context, tpl template.Resource, variables *PreparedTemplateVariables) (*template.RenderResult, error) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)

	return s.renderTemplateInLanguage(preferredLanguageTags, tpl, variables)
}

func (s *Service) renderTemplateInLanguage(preferredLanguages []string, tpl template.Resource, variables *PreparedTemplateVariables) (*template.RenderResult, error) {
	out, err := s.TemplateEngine.Render(tpl, preferredLanguages, variables)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Service) GetSenderForTestEmail(ctx context.Context) (sender string, err error) {
	t, err := s.translationMap(ctx)
	if err != nil {
		return
	}

	sender, err = t.RenderText("email.default.sender", nil)
	if err != nil {
		return
	}

	return
}

func (s *Service) GetSenderForTestSMS(ctx context.Context) (sender string, err error) {
	return s.smsMessageHeader(ctx, "default", &PreparedTemplateVariables{})
}

func (s *Service) emailMessageHeader(ctx context.Context, name SpecName, variables *PreparedTemplateVariables) (sender, replyTo, subject string, err error) {
	t, err := s.translationMap(ctx)
	if err != nil {
		return
	}

	sender, err = t.RenderText(fmt.Sprintf("email.%s.sender", name), variables)
	if errors.Is(err, template.ErrNotFound) {
		sender, err = t.RenderText("email.default.sender", variables)
	}
	if err != nil {
		return
	}

	replyTo, err = t.RenderText(fmt.Sprintf("email.%s.reply-to", name), variables)
	if errors.Is(err, template.ErrNotFound) {
		replyTo, err = t.RenderText("email.default.reply-to", variables)
	}
	if err != nil {
		return
	}

	subject, err = t.RenderText(fmt.Sprintf("email.%s.subject", name), variables)
	if errors.Is(err, template.ErrNotFound) {
		subject, err = t.RenderText("email.default.subject", variables)
	}
	if err != nil {
		return
	}

	return
}

func (s *Service) EmailMessageData(ctx context.Context, msg *MessageSpec, variables *PartialTemplateVariables) (*EmailMessageData, error) {
	// Ensure these data are safe to put at query
	textData, err := s.prepareTemplateVariables(ctx, variables)
	if err != nil {
		return nil, err
	}
	textData.ClientID = htmltemplate.URLQueryEscaper(textData.ClientID)
	textData.State = htmltemplate.URLQueryEscaper(textData.State)
	textData.XState = htmltemplate.URLQueryEscaper(textData.XState)
	textData.UILocales = htmltemplate.URLQueryEscaper(textData.UILocales)

	// html template will handle the escape
	htmlData, err := s.prepareTemplateVariables(ctx, variables)
	if err != nil {
		return nil, err
	}

	sender, replyTo, subject, err := s.emailMessageHeader(ctx, msg.Name, htmlData)
	if err != nil {
		return nil, err
	}

	textBody, err := s.renderTemplate(ctx, msg.TXTEmailTemplate, textData)
	if err != nil {
		return nil, err
	}

	htmlBody, err := s.renderTemplate(ctx, msg.HTMLEmailTemplate, htmlData)
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

func (s *Service) smsMessageHeader(ctx context.Context, name SpecName, variables *PreparedTemplateVariables) (sender string, err error) {
	t, err := s.translationMap(ctx)
	if err != nil {
		return
	}

	sender, err = t.RenderText(fmt.Sprintf("sms.%s.sender", name), variables)
	if errors.Is(err, template.ErrNotFound) {
		sender, err = t.RenderText("sms.default.sender", variables)
	}
	if err != nil {
		return
	}

	return
}

func (s *Service) SMSMessageData(ctx context.Context, msg *MessageSpec, variables *PartialTemplateVariables) (*SMSMessageData, error) {
	data, err := s.prepareTemplateVariables(ctx, variables)
	if err != nil {
		return nil, err
	}
	data.ClientID = htmltemplate.URLQueryEscaper(data.ClientID)
	data.State = htmltemplate.URLQueryEscaper(data.State)
	data.XState = htmltemplate.URLQueryEscaper(data.XState)
	data.UILocales = htmltemplate.URLQueryEscaper(data.UILocales)

	sender, err := s.smsMessageHeader(ctx, msg.Name, data)
	if err != nil {
		return nil, err
	}

	body, err := s.renderTemplate(ctx, msg.SMSTemplate, data)
	if err != nil {
		return nil, err
	}

	return &SMSMessageData{
		Sender:                    sender,
		Body:                      body,
		PreparedTemplateVariables: data,
	}, nil
}

func (s *Service) WhatsappMessageData(ctx context.Context, language string, msg *MessageSpec, variables *PartialTemplateVariables) (*WhatsappMessageData, error) {
	data, err := s.prepareTemplateVariables(ctx, variables)
	if err != nil {
		return nil, err
	}

	body, err := s.renderTemplateInLanguage([]string{language}, msg.WhatsappTemplate, data)
	if err != nil {
		return nil, err
	}

	return &WhatsappMessageData{
		Body: body,
	}, nil
}

func (s *Service) HasKey(ctx context.Context, key string) (bool, error) {
	t, err := s.translationMap(ctx)
	if err != nil {
		return false, err
	}
	return t.HasKey(key), nil
}

func (s *Service) RenderText(ctx context.Context, key string, args interface{}) (string, error) {
	t, err := s.translationMap(ctx)
	if err != nil {
		return "", err
	}
	return t.RenderText(key, args)
}

func (s *Service) prepareTemplateVariables(ctx context.Context, v *PartialTemplateVariables) (*PreparedTemplateVariables, error) {
	t, err := s.translationMap(ctx)
	if err != nil {
		return nil, err
	}

	// TODO(l10n): investigate on how to allow referencing other translation natively.
	appName, err := t.RenderText("app.name", nil)
	if err != nil {
		return nil, err
	}

	uiParams := uiparam.GetUIParam(ctx)

	return &PreparedTemplateVariables{
		AppName:     appName,
		ClientID:    uiParams.ClientID,
		Code:        v.Code,
		Email:       v.Email,
		HasPassword: v.HasPassword,
		Host:        v.Host,
		Link:        v.Link,
		Password:    v.Password,
		Phone:       v.Phone,
		State:       uiParams.State,
		StaticAssetURL: func(id string) (url string, err error) {
			return s.StaticAssets.StaticAssetURL(ctx, id)
		},
		UILocales: uiParams.UILocales,
		URL:       v.URL,
		XState:    uiParams.XState,
	}, nil
}
