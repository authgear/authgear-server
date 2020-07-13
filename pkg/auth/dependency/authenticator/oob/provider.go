package oob

import (
	"context"
	"errors"
	"net/url"
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	taskspec "github.com/authgear/authgear-server/pkg/auth/task/spec"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/intl"
	"github.com/authgear/authgear-server/pkg/core/uuid"
	"github.com/authgear/authgear-server/pkg/mail"
	"github.com/authgear/authgear-server/pkg/otp"
	"github.com/authgear/authgear-server/pkg/sms"
	"github.com/authgear/authgear-server/pkg/task"
	"github.com/authgear/authgear-server/pkg/template"
)

type EndpointsProvider interface {
	BaseURL() *url.URL
}

type Provider struct {
	Context        context.Context
	Localization   *config.LocalizationConfig
	AppMetadata    config.AppMetadata
	Messaging      *config.MessagingConfig
	Config         *config.AuthenticatorOOBConfig
	Store          *Store
	TemplateEngine *template.Engine
	Endpoints      EndpointsProvider
	TaskQueue      task.Queue
	Clock          clock.Clock
}

func (p *Provider) Get(userID string, id string) (*Authenticator, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByChannel(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string) (*Authenticator, error) {
	return p.Store.GetByChannel(userID, channel, phone, email)
}

func (p *Provider) Delete(a *Authenticator) error {
	return p.Store.Delete(a.ID)
}

func (p *Provider) List(userID string) ([]*Authenticator, error) {
	authenticators, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortAuthenticators(authenticators)
	return authenticators, nil
}

func (p *Provider) New(userID string, channel authn.AuthenticatorOOBChannel, phone string, email string) *Authenticator {
	a := &Authenticator{
		ID:      uuid.New(),
		UserID:  userID,
		Channel: channel,
		Phone:   phone,
		Email:   email,
	}
	return a
}

func (p *Provider) Create(a *Authenticator) error {
	_, err := p.Store.GetByChannel(a.UserID, a.Channel, a.Phone, a.Email)
	if err == nil {
		return authenticator.ErrAuthenticatorAlreadyExists
	} else if !errors.Is(err, authenticator.ErrAuthenticatorNotFound) {
		return err
	}

	now := p.Clock.NowUTC()
	a.CreatedAt = now

	return p.Store.Create(a)
}

func (p *Provider) Authenticate(expectedCode string, code string) error {
	ok := otp.ValidateOOBOTP(expectedCode, code)
	if !ok {
		return errors.New("invalid bearer token")
	}
	return nil
}

func (p *Provider) GenerateCode() string {
	return otp.GenerateOOBOTP()
}

type SendCodeOptions struct {
	Channel string
	Email   string
	Phone   string
	Code    string
}

func (p *Provider) SendCode(opts SendCodeOptions) (err error) {
	email := opts.Email
	phone := opts.Phone
	channel := opts.Channel
	code := opts.Code

	data := map[string]interface{}{
		"email": email,
		"phone": phone,
		"code":  code,
		"host":  p.Endpoints.BaseURL().Host,
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(p.Context)
	data["appname"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.Localization.FallbackLanguage), p.AppMetadata, "app_name")

	switch channel {
	case string(authn.AuthenticatorOOBChannelEmail):
		return p.SendEmail(email, data)
	case string(authn.AuthenticatorOOBChannelSMS):
		return p.SendSMS(phone, data)
	default:
		panic("expected OOB channel: " + string(channel))
	}
}

func (p *Provider) SendEmail(email string, data map[string]interface{}) (err error) {
	textBody, err := p.TemplateEngine.RenderTemplate(
		otp.TemplateItemTypeOOBCodeEmailTXT,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	htmlBody, err := p.TemplateEngine.RenderTemplate(
		otp.TemplateItemTypeOOBCodeEmailHTML,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(task.Spec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			EmailMessages: []mail.SendOptions{
				{
					MessageConfig: config.NewEmailMessageConfig(
						p.Messaging.DefaultEmailMessage,
						p.Config.Email.Message,
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

func (p *Provider) SendSMS(phone string, data map[string]interface{}) (err error) {
	body, err := p.TemplateEngine.RenderTemplate(
		otp.TemplateItemTypeOOBCodeSMSTXT,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(task.Spec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			SMSMessages: []sms.SendOptions{
				{
					MessageConfig: config.NewSMSMessageConfig(
						p.Messaging.DefaultSMSMessage,
						p.Config.SMS.Message,
					),
					To:   phone,
					Body: body,
				},
			},
		},
	})

	return
}
func sortAuthenticators(as []*Authenticator) {
	sort.Slice(as, func(i, j int) bool {
		return as[i].CreatedAt.Before(as[j].CreatedAt)
	})
}
