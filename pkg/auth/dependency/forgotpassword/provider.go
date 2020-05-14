package forgotpassword

import (
	"net/url"
	"path"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	taskspec "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
	coretime "github.com/skygeario/skygear-server/pkg/core/time"
)

type ResetPasswordFlow interface {
	ResetPassword(userID string, password string) error
}

type LoginIDProvider interface {
	GetByLoginID(loginID loginid.LoginID) ([]*loginid.Identity, error)
	IsLoginIDKeyType(loginIDKey string, loginIDKeyType metadata.StandardKey) bool
}

type Provider struct {
	StaticAssetURLPrefix        string
	AppName                     string
	EmailMessageConfiguration   config.EmailMessageConfiguration
	SMSMessageConfiguration     config.SMSMessageConfiguration
	ForgotPasswordConfiguration *config.ForgotPasswordConfiguration

	Store Store

	AuthInfoStore     authinfo.Store
	UserProfileStore  userprofile.Store
	HookProvider      hook.Provider
	TimeProvider      coretime.Provider
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
	createdAt := p.TimeProvider.NowUTC()
	codeStr = GenerateCode()
	expireAt := createdAt.Add(time.Duration(p.ForgotPasswordConfiguration.ResetCodeLifetime) * time.Second)
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
		"appname":                 p.AppName,
		"email":                   email,
		"code":                    code,
		"link":                    u.String(),
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

	p.TaskQueue.Enqueue(async.TaskSpec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			EmailMessages: []mail.SendOptions{
				mail.SendOptions{
					MessageConfig: config.NewEmailMessageConfiguration(
						p.EmailMessageConfiguration,
						p.ForgotPasswordConfiguration.EmailMessage,
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

	p.TaskQueue.Enqueue(async.TaskSpec{
		Name: taskspec.SendMessagesTaskName,
		Param: taskspec.SendMessagesTaskParam{
			SMSMessages: []sms.SendOptions{
				sms.SendOptions{
					MessageConfig: config.NewSMSMessageConfiguration(
						p.SMSMessageConfiguration,
						p.ForgotPasswordConfiguration.SMSMessage,
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

	now := p.TimeProvider.NowUTC()
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

	var authInfo authinfo.AuthInfo
	err = p.AuthInfoStore.GetAuth(userID, &authInfo)
	if err != nil {
		return
	}

	userProfile, err := p.UserProfileStore.GetUserProfile(userID)
	if err != nil {
		return
	}

	user := model.NewUser(authInfo, userProfile)

	err = p.HookProvider.DispatchEvent(
		event.PasswordUpdateEvent{
			Reason: event.PasswordUpdateReasonResetPassword,
			User:   user,
		},
		&user,
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
