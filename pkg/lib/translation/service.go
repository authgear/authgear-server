package translation

import (
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package translation_test

type StaticAssetResolver interface {
	StaticAssetURL(ctx context.Context, id string) (url string, err error)
}

type Service struct {
	TemplateEngine                  *template.Engine
	StaticAssets                    StaticAssetResolver
	SMTPServerCredentialsSecretItem *config.SMTPServerCredentialsSecretItem
	OAuthConfig                     *config.OAuthConfig

	translations *template.TranslationMap `wire:"-"`
}

func (s *Service) translationMap(ctx context.Context) (*template.TranslationMap, error) {
	if s.translations == nil {
		preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)
		t, err := s.TemplateEngine.Translation(ctx, preferredLanguageTags)
		if err != nil {
			return nil, err
		}
		s.translations = t
	}
	return s.translations, nil
}

func (s *Service) levelSpecificTranslationMap(ctx context.Context, level resource.FsLevel) (*template.TranslationMap, error) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)
	t, err := s.TemplateEngine.LevelSpecificTranslation(ctx, level, preferredLanguageTags)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) renderTemplate(ctx context.Context, tpl template.Resource, variables *PreparedTemplateVariables) (*template.RenderResult, error) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(ctx)

	return s.renderTemplateInLanguage(ctx, preferredLanguageTags, tpl, variables)
}

func (s *Service) renderTemplateInLanguage(ctx context.Context, preferredLanguages []string, tpl template.Resource, variables *PreparedTemplateVariables) (*template.RenderResult, error) {
	out, err := s.TemplateEngine.Render(ctx, tpl, preferredLanguages, variables)
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
	effectiveTranslations, err := s.translationMap(ctx)
	if err != nil {
		return
	}

	resolveSender := func(t *template.TranslationMap) (string, error) {
		sender, err := t.RenderText(fmt.Sprintf("email.%s.sender", name), variables)
		if errors.Is(err, template.ErrNotFound) {
			sender, err = t.RenderText("email.default.sender", variables)
		}
		if err != nil {
			return "", err
		}
		return sender, nil
	}

	// Resolve sender
	// If no smtp secret, probably in local, just use sender in translation
	if s.SMTPServerCredentialsSecretItem == nil {
		sender, err = resolveSender(effectiveTranslations)
		if err != nil {
			return
		}
	} else {
		// If the secret has sender, use it.
		// If the developer wants to have different senders for different locales,
		// they have to remove the sender in the secret.
		if s.SMTPServerCredentialsSecretItem.GetData().Sender != "" {
			sender = s.SMTPServerCredentialsSecretItem.GetData().Sender
		} else {
			// Depends on the secret fs level, resolve sender from different level of translation
			var levelTranslations *template.TranslationMap
			levelTranslations, err = s.levelSpecificTranslationMap(ctx, s.SMTPServerCredentialsSecretItem.FsLevel)
			if err != nil {
				return
			}
			sender, err = resolveSender(levelTranslations)
			if err != nil {
				return
			}
		}
	}

	replyTo, err = effectiveTranslations.RenderText(fmt.Sprintf("email.%s.reply-to", name), variables)
	if errors.Is(err, template.ErrNotFound) {
		replyTo, err = effectiveTranslations.RenderText("email.default.reply-to", variables)
	}
	if err != nil {
		return
	}

	subject, err = effectiveTranslations.RenderText(fmt.Sprintf("email.%s.subject", name), variables)
	if errors.Is(err, template.ErrNotFound) {
		subject, err = effectiveTranslations.RenderText("email.default.subject", variables)
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

	body, err := s.renderTemplateInLanguage(ctx, []string{language}, msg.WhatsappTemplate, data)
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
	clientName := ""
	if uiParams.ClientID != "" {
		if client, ok := s.OAuthConfig.GetClient(uiParams.ClientID); ok {
			clientName = client.Name
		}
	}

	return &PreparedTemplateVariables{
		AppName:     appName,
		ClientID:    uiParams.ClientID,
		ClientName:  clientName,
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
