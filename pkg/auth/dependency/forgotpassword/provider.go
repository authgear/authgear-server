package forgotpassword

import (
	"context"
	"net/url"
	"path"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	taskspec "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/intl"
	"github.com/skygeario/skygear-server/pkg/mail"
	"github.com/skygeario/skygear-server/pkg/sms"
	"github.com/skygeario/skygear-server/pkg/template"
)

type ResetPasswordFlow interface {
	ResetPassword(userID string, password string) error
}

type LoginIDProvider interface {
	GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error)
	IsLoginIDKeyType(loginIDKey string, loginIDKeyType metadata.StandardKey) bool
}

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type Provider struct {
	Context               context.Context
	StaticAssetURLPrefix  string
	LocalizationConfig    *config.LocalizationConfig
	MetadataConfiguration config.AppMetadata
	MessagingConfig       config.MessagingConfig
	ForgotPasswordConfig  *config.ForgotPasswordConfig

	Store Store

	Users             UserProvider
	HookProvider      hook.Provider
	Clock             clock.Clock
	URLPrefixProvider urlprefix.Provider
	TemplateEngine    *template.Engine
	TaskQueue         async.Queue

	Interactions    ResetPasswordFlow
	LoginIDProvider LoginIDProvider
}

// SendCode checks if loginID is an existing login ID.
// For first matched login ID, a code is generated.
// Other matched login IDs are ignored.
// The code expires after a specific time.
// The code becomes invalid if it is consumed.
// Finally the code is sent to the login ID asynchronously.
func (p *Provider) SendCode(loginID string) (err error) {
	// TODO(forgotpassword): Test SendCode
	idens, err := p.LoginIDProvider.GetByLoginID(
		loginid.LoginID{
			Key:   "",
			Value: loginID,
		},
	)
	if err != nil {
		return
	}

	for _, iden := range idens {
		email := p.LoginIDProvider.IsLoginIDKeyType(iden.LoginIDKey, metadata.Email)
		phone := p.LoginIDProvider.IsLoginIDKeyType(iden.LoginIDKey, metadata.Phone)

		if !email && !phone {
			continue
		}

		code, codeStr := p.newCode(iden.UserID)

		err = p.Store.Create(code)
		if err != nil {
			return
		}

		if email {
			err = p.sendEmail(iden.LoginID, codeStr)
			return
		}

		if phone {
			err = p.sendSMS(iden.LoginID, codeStr)
			return
		}
	}

	return
}

func (p *Provider) newCode(userID string) (code *Code, codeStr string) {
	createdAt := p.Clock.NowUTC()
	codeStr = GenerateCode()
	expireAt := createdAt.Add(time.Duration(p.ForgotPasswordConfig.ResetCodeExpiry) * time.Second)
	code = &Code{
		CodeHash:  HashCode(codeStr),
		UserID:    userID,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		Consumed:  false,
	}
	return
}

func (p *Provider) sendEmail(email string, code string) (err error) {
	u := p.makeURL(code)

	data := map[string]interface{}{
		"static_asset_url_prefix": p.StaticAssetURLPrefix,
		"email":                   email,
		"code":                    code,
		"link":                    u.String(),
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(p.Context)
	data["appname"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.LocalizationConfig.FallbackLanguage), p.MetadataConfiguration, "app_name")

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

	p.TaskQueue.Enqueue(async.TaskSpec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			EmailMessages: []mail.SendOptions{
				mail.SendOptions{
					MessageConfig: config.NewEmailMessageConfig(
						p.MessagingConfig.DefaultEmailMessage,
						p.ForgotPasswordConfig.EmailMessage,
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

func (p *Provider) sendSMS(phone string, code string) (err error) {
	u := p.makeURL(code)

	data := map[string]interface{}{
		"code": code,
		"link": u.String(),
	}

	preferredLanguageTags := intl.GetPreferredLanguageTags(p.Context)
	data["appname"] = intl.LocalizeJSONObject(preferredLanguageTags, intl.Fallback(p.LocalizationConfig.FallbackLanguage), p.MetadataConfiguration, "app_name")

	body, err := p.TemplateEngine.RenderTemplate(
		TemplateItemTypeForgotPasswordSMSTXT,
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
				sms.SendOptions{
					MessageConfig: config.NewSMSMessageConfig(
						p.MessagingConfig.DefaultSMSMessage,
						p.ForgotPasswordConfig.SMSMessage,
					),
					To:   phone,
					Body: body,
				},
			},
		},
	})

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
func (p *Provider) ResetPassword(codeStr string, newPassword string) (err error) {
	codeHash := HashCode(codeStr)
	code, err := p.Store.Get(codeHash)
	if err != nil {
		return
	}

	now := p.Clock.NowUTC()
	if now.After(code.ExpireAt) {
		err = ErrExpiredCode
		return
	}
	if code.Consumed {
		err = ErrUsedCode
		return
	}

	userID := code.UserID
	err = p.Interactions.ResetPassword(userID, newPassword)
	if err != nil {
		return err
	}

	user, err := p.Users.Get(userID)
	if err != nil {
		return
	}

	err = p.HookProvider.DispatchEvent(
		event.PasswordUpdateEvent{
			Reason: event.PasswordUpdateReasonResetPassword,
			User:   *user,
		},
		user,
	)
	if err != nil {
		return
	}

	// We have to mark the code as consumed at the end
	// because if we mark it at the beginning,
	// the code will be consumed if the new password violates
	// the password policy.
	code.Consumed = true
	err = p.Store.Update(code)
	if err != nil {
		return
	}

	p.TaskQueue.Enqueue(async.TaskSpec{
		Name: taskspec.PwHousekeeperTaskName,
		Param: taskspec.PwHousekeeperTaskParam{
			AuthID: user.ID,
		},
	})

	return nil
}
