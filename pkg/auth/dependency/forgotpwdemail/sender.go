package forgotpwdemail

import (
	"fmt"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type Sender interface {
	Send(
		email string,
		authInfo authinfo.AuthInfo,
		userProfile userprofile.UserProfile,
		hashedPassword []byte,
	) error
}

type DefaultSender struct {
	Config         config.ForgotPasswordConfiguration
	Dialer         *gomail.Dialer
	CodeGenerator  *CodeGenerator
	TemplateEngine *template.Engine
}

func NewDefaultSender(
	config config.TenantConfiguration,
	dialer *gomail.Dialer,
	templateEngine *template.Engine,
) Sender {
	return &DefaultSender{
		Config:         config.ForgotPassword,
		Dialer:         dialer,
		CodeGenerator:  &CodeGenerator{config.MasterKey},
		TemplateEngine: templateEngine,
	}
}

func (d *DefaultSender) Send(
	email string,
	authInfo authinfo.AuthInfo,
	userProfile userprofile.UserProfile,
	hashedPassword []byte,
) (err error) {
	expireAt :=
		time.Now().UTC().
			Truncate(time.Second * 1).
			Add(time.Second * time.Duration(d.Config.ResetURLLifeTime))
	code := d.CodeGenerator.Generate(authInfo, email, hashedPassword, expireAt)
	context := map[string]interface{}{
		"appname": d.Config.AppName,
		"link": fmt.Sprintf(
			"%s/forgot_password/reset_password_form?code=%s&user_id=%s&expire_at=%d",
			d.Config.URLPrefix,
			code,
			authInfo.ID,
			expireAt.UTC().Unix(),
		),
		"email":      email,
		"user_id":    userProfile.ID,
		"user":       userProfile,
		"url_prefix": d.Config.URLPrefix,
		"code":       code,
		"expire_at":  expireAt,
	}

	var textBody string
	if textBody, err = d.TemplateEngine.ParseTextTemplate(
		authTemplate.TemplateNameForgotPasswordEmailText,
		context,
		template.ParseOption{Required: true},
	); err != nil {
		return
	}

	var htmlBody string
	if htmlBody, err = d.TemplateEngine.ParseHTMLTemplate(
		authTemplate.TemplateNameForgotPasswordEmailHTML,
		context,
		template.ParseOption{Required: false},
	); err != nil {
		return
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
