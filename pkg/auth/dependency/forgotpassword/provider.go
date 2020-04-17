package forgotpassword

import (
	"net/url"
	"path"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

type Provider struct {
	AppName                     string
	EmailMessageConfiguration   config.EmailMessageConfiguration
	ForgotPasswordConfiguration *config.ForgotPasswordConfiguration
	PasswordAuthProvider        password.Provider
	Store                       Store
	TimeProvider                coretime.Provider
	URLPrefixProvider           urlprefix.Provider
	TemplateEngine              *template.Engine
	MailSender                  mail.Sender
	SMSClient                   sms.Client
}

// SendCode checks if loginID is an existing login ID.
// If not found, ErrLoginIDNotFound is returned.
// If the login ID is not of type email or phone, ErrUnsupportedLoginIDType is returned.
// Otherwise, a code is generated.
// The code expires after a specific time.
// The code becomes invalid if it is consumed.
// Finally the code is sent to the login ID asynchronously.
func (p *Provider) SendCode(loginID string) (err error) {
	// TODO(forgotpassword): Test SendCode
	// Send single email
	// Send multiple email
	// Send email with the normalized email
	prins, err := p.PasswordAuthProvider.GetPrincipalsByLoginID("", loginID)
	if err != nil {
		return
	}

	if len(prins) <= 0 {
		err = ErrLoginIDNotFound
		return
	}

	for _, prin := range prins {
		email := p.PasswordAuthProvider.CheckLoginIDKeyType(prin.LoginIDKey, metadata.Email)
		phone := p.PasswordAuthProvider.CheckLoginIDKeyType(prin.LoginIDKey, metadata.Phone)

		if !email && !phone {
			err = ErrUnsupportedLoginIDType
			return
		}

		code, codeStr := p.newCode(prin)

		err = p.Store.StoreCode(code)
		if err != nil {
			return
		}

		if email {
			err = p.sendEmail(prin.LoginID, codeStr)
			return
		}

		if phone {
			err = p.sendSMS(prin.LoginID, codeStr)
			return
		}
	}

	return
}

func (p *Provider) newCode(prin *password.Principal) (code *Code, codeStr string) {
	createdAt := p.TimeProvider.NowUTC()
	codeStr = GenerateCode()
	expireAt := createdAt.Add(time.Duration(p.ForgotPasswordConfiguration.ResetURLLifetime) * time.Second)
	code = &Code{
		CodeHash:    HashCode(codeStr),
		PrincipalID: prin.ID,
		CreatedAt:   createdAt,
		ExpireAt:    expireAt,
		Consumed:    false,
	}
	return
}

func (p *Provider) sendEmail(email string, code string) (err error) {
	u := p.makeURL(code)

	data := map[string]interface{}{
		"appname": p.AppName,
		"email":   email,
		"code":    code,
		"link":    u.String(),
	}

	textBody, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeForgotPasswordEmailTXT,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	htmlBody, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeForgotPasswordEmailHTML,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	messageConfig := config.NewEmailMessageConfiguration(
		p.EmailMessageConfiguration,
		p.ForgotPasswordConfiguration.EmailMessage,
	)
	err = p.MailSender.Send(mail.SendOptions{
		MessageConfig: messageConfig,
		Recipient:     email,
		TextBody:      textBody,
		HTMLBody:      htmlBody,
	})
	if err != nil {
		return
	}

	return
}

func (p *Provider) sendSMS(phone string, code string) (err error) {
	u := p.makeURL(code)

	data := map[string]interface{}{
		"appname": p.AppName,
		"code":    code,
		"link":    u.String(),
	}

	body, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeForgotPasswordSMSTXT,
		data,
		template.ResolveOptions{},
	)
	if err != nil {
		return
	}

	err = p.SMSClient.Send(sms.SendOptions{
		// TODO(forgotpassword): support SMS message config
		MessageConfig: nil,
		To:            phone,
		Body:          body,
	})
	if err != nil {
		return
	}

	return
}

func (p *Provider) makeURL(code string) *url.URL {
	u := *p.URLPrefixProvider.Value()
	// /reset_password is an endpoint of Auth UI.
	u.Path = path.Join(u.Path, "reset_password")
	u.RawQuery = url.Values{
		"code": []string{code},
	}.Encode()
	return &u
}

// ResetPassword consumes code and reset password to newPassword.
// If the code is invalid, ErrInvalidCode is returned.
// If the code is found but expired, ErrExpiredCode is returned.
// if the code is found but used, ErrUsedCode is returned.
// Otherwise, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (p *Provider) ResetPassword(code string, newPassword string) error {
	// TODO(forgotpassword)
	return nil
}
