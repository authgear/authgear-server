package welcemail

import (
	"errors"

	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type Sender interface {
	Send(email string, userProfile userprofile.UserProfile) error
}

type DefaultSender struct {
	AppName string
	Config  config.WelcomeEmailConfiguration
	Dialer  *gomail.Dialer
}

func NewDefaultSender(config config.TenantConfiguration, dialer *gomail.Dialer) Sender {
	return &DefaultSender{
		AppName: config.AppName,
		Config:  config.WelcomeEmail,
		Dialer:  dialer,
	}
}

func (d *DefaultSender) Send(email string, userProfile userprofile.UserProfile) (err error) {
	if d.Config.TextURL == "" {
		return errors.New("welcome email text template url is empty")
	}

	context := map[string]interface{}{
		"appname": d.AppName,
		"email":   email,
		"user_id": userProfile.ID,
		"user":    userProfile.ToMap(),
		// TODO: url prefix
	}

	var textBody string
	if textBody, err = template.ParseTextTemplateFromURL(d.Config.TextURL, context); err != nil {
		return
	}

	var htmlBody string
	if d.Config.HTMLURL != "" {
		if htmlBody, err = template.ParseHTMLTemplateFromURL(d.Config.HTMLURL, context); err != nil {
			return
		}
	}

	sendReq := mail.SendRequest{
		Dialer:      d.Dialer,
		Sender:      d.Config.Sender,
		SenderName:  d.Config.SenderName,
		Recipient:   email,
		Subject:     d.Config.Subject,
		ReplyTo:     d.Config.ReplyTo,
		ReplyToName: d.Config.ReplyToName,
		TextBody:    textBody,
		HTMLBody:    htmlBody,
	}

	err = sendReq.Execute()
	return
}
