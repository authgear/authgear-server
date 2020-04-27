package interaction

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	taskspec "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type OOBProviderImpl struct {
	SMSMessageConfiguration       config.SMSMessageConfiguration
	EmailMessageConfiguration     config.EmailMessageConfiguration
	AuthenticatorOOBConfiguration *config.AuthenticatorOOBConfiguration
	TemplateEngine                *template.Engine
	URLPrefixProvider             urlprefix.Provider
	TaskQueue                     async.Queue
}

func (p *OOBProviderImpl) GenerateCode() string {
	return oob.GenerateCode()
}

func (p *OOBProviderImpl) SendCode(spec AuthenticatorSpec, code string) (err error) {
	urlPrefix := p.URLPrefixProvider.Value()
	email := ""
	phone := ""
	if s, ok := spec.Props[AuthenticatorPropOOBOTPEmail].(string); ok {
		email = s
	}
	if s, ok := spec.Props[AuthenticatorPropOOBOTPPhone].(string); ok {
		phone = s
	}
	channel := spec.Props[AuthenticatorPropOOBOTPChannelType].(string)

	data := map[string]interface{}{
		"email": email,
		"phone": phone,
		"code":  code,
		"host":  urlPrefix.Host,
	}

	switch channel {
	case string(authn.AuthenticatorOOBChannelEmail):
		return p.SendEmail(email, data)
	case string(authn.AuthenticatorOOBChannelSMS):
		return p.SendSMS(phone, data)
	default:
		panic("expected OOB channel: " + string(channel))
	}
}

func (p *OOBProviderImpl) SendEmail(email string, data map[string]interface{}) (err error) {
	textBody, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeOOBCodeEmailTXT,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	htmlBody, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeOOBCodeEmailHTML,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(async.TaskSpec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			EmailMessages: []mail.SendOptions{
				{
					MessageConfig: config.NewEmailMessageConfiguration(
						p.EmailMessageConfiguration,
						p.AuthenticatorOOBConfiguration.Email.Message,
					),
					Recipient: email,
					TextBody:  textBody,
					HTMLBody:  htmlBody,
				},
			},
		},
	})

	return
}

func (p *OOBProviderImpl) SendSMS(phone string, data map[string]interface{}) (err error) {
	body, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeOOBCodeSMSTXT,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(async.TaskSpec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			SMSMessages: []sms.SendOptions{
				{
					MessageConfig: config.NewSMSMessageConfiguration(
						p.SMSMessageConfiguration,
						p.AuthenticatorOOBConfiguration.SMS.Message,
					),
					To:   phone,
					Body: body,
				},
			},
		},
	})

	return
}
