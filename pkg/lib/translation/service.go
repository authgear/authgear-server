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
}

func (t *Service) renderTranslation(key string, args interface{}) (string, error) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(t.Context)
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

	out, err := t.TemplateEngine.RenderTranslation(
		renderCtx,
		TemplateItemTypeTranslationJSON,
		key,
		args,
	)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (t *Service) renderTemplate(typ string, args interface{}) (string, error) {
	preferredLanguageTags := intl.GetPreferredLanguageTags(t.Context)

	renderCtx := &template.RenderContext{
		PreferredLanguageTags: preferredLanguageTags,
	}

	out, err := t.TemplateEngine.Render(renderCtx, typ, args)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (t *Service) AppMetadata() (*AppMetadata, error) {
	args := map[string]interface{}{
		"StaticAssetURLPrefix": t.EnvironmentConfig.StaticAssetURLPrefix,
	}

	appName, err := t.renderTranslation("app.app-name", args)
	if err != nil {
		return nil, err
	}

	logoURI, err := t.renderTranslation("app.logo-uri", args)
	if err != nil {
		return nil, err
	}

	return &AppMetadata{
		AppName: appName,
		LogoURI: logoURI,
	}, nil
}

func (t *Service) emailMessageHeader(name string, args interface{}) (sender, replyTo, subject string, err error) {
	sender, err = t.renderTranslation(fmt.Sprintf("email.%s.sender", name), args)
	if template.IsNotFound(err) {
		sender, err = t.renderTranslation("email.default.sender", args)
	}
	if err != nil {
		return
	}

	replyTo, err = t.renderTranslation(fmt.Sprintf("email.%s.reply-to", name), args)
	if template.IsNotFound(err) {
		replyTo, err = t.renderTranslation("email.default.reply-to", args)
	}
	if err != nil {
		return
	}

	subject, err = t.renderTranslation(fmt.Sprintf("email.%s.subject", name), args)
	if template.IsNotFound(err) {
		subject, err = t.renderTranslation("email.default.subject", args)
	}
	if err != nil {
		return
	}

	return
}

func (t *Service) EmailMessageData(msg *MessageSpec, args interface{}) (*EmailMessageData, error) {
	sender, replyTo, subject, err := t.emailMessageHeader(msg.Name, args)
	if err != nil {
		return nil, err
	}

	textBody, err := t.renderTemplate(msg.TXTEmailType, args)
	if err != nil {
		return nil, err
	}

	htmlBody, err := t.renderTemplate(msg.HTMLEmailType, args)
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

func (t *Service) smsMessageHeader(name string, args interface{}) (sender string, err error) {
	sender, err = t.renderTranslation(fmt.Sprintf("sms.%s.sender", name), args)
	if template.IsNotFound(err) {
		sender, err = t.renderTranslation("sms.default.sender", args)
	}
	if err != nil {
		return
	}

	return
}

func (t *Service) SMSMessageData(msg *MessageSpec, args interface{}) (*SMSMessageData, error) {
	sender, err := t.smsMessageHeader(msg.Name, args)
	if err != nil {
		return nil, err
	}

	body, err := t.renderTemplate(msg.SMSType, args)
	if err != nil {
		return nil, err
	}

	return &SMSMessageData{
		Sender: sender,
		Body:   body,
	}, nil
}
