package forgotpassword

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type TemplateData struct {
	Email string
	Code  string
	Link  string
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec, secret string) (*authenticator.Info, error)
	WithSecret(ai *authenticator.Info, secret string) (bool, *authenticator.Info, error)
}

type IdentityService interface {
	ListByClaim(name string, value string) ([]*identity.Info, error)
}

type URLProvider interface {
	ResetPasswordURL(code string) *url.URL
}

type TranslationService interface {
	EmailMessageData(msg *translation.MessageSpec, args interface{}) (*translation.EmailMessageData, error)
	SMSMessageData(msg *translation.MessageSpec, args interface{}) (*translation.SMSMessageData, error)
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type ProviderLogger struct{ *log.Logger }

func NewProviderLogger(lf *log.Factory) ProviderLogger {
	return ProviderLogger{lf.New("forgotpassword")}
}

type Provider struct {
	Request     *http.Request
	Translation TranslationService
	Config      *config.ForgotPasswordConfig
	TrustProxy  config.TrustProxy

	Store     *Store
	Clock     clock.Clock
	URLs      URLProvider
	TaskQueue task.Queue

	Logger ProviderLogger

	Identities     IdentityService
	Authenticators AuthenticatorService
	RateLimiter    RateLimiter
}

// SendCode checks if loginID is an existing login ID.
// For first matched login ID, a code is generated.
// Other matched login IDs are ignored.
// The code expires after a specific time.
// The code becomes invalid if it is consumed.
// Finally the code is sent to the login ID asynchronously.
func (p *Provider) SendCode(loginID string) error {
	err := p.RateLimiter.TakeToken(SendResetPasswordCodeRateLimitBucket(loginID))
	if err != nil {
		return err
	}

	emailIdentities, err := p.Identities.ListByClaim(string(authn.ClaimEmail), loginID)
	if err != nil {
		return err
	}
	phoneIdentities, err := p.Identities.ListByClaim(string(authn.ClaimPhoneNumber), loginID)
	if err != nil {
		return err
	}

	allIdentities := append(emailIdentities, phoneIdentities...)
	if len(allIdentities) == 0 {
		return ErrUserNotFound
	}

	for _, info := range allIdentities {
		authenticators, err := p.Authenticators.List(
			info.UserID,
			authenticator.KeepType(authn.AuthenticatorTypePassword),
			authenticator.KeepKind(authenticator.KindPrimary),
		)
		if err != nil {
			return err
		} else if len(authenticators) == 0 {
			return ErrNoPassword
		}
	}

	for _, info := range emailIdentities {
		email := info.Claims[string(authn.ClaimEmail)].(string)
		code, codeStr := p.newCode(info.UserID)

		if err := p.Store.Create(code); err != nil {
			return err
		}

		p.Logger.Debugf("sending email")
		if err := p.sendEmail(email, codeStr); err != nil {
			return err
		}
	}

	for _, info := range phoneIdentities {
		phone := info.Claims[string(authn.ClaimPhoneNumber)].(string)
		code, codeStr := p.newCode(info.UserID)

		if err := p.Store.Create(code); err != nil {
			return err
		}

		p.Logger.Debugf("sending sms")
		if err := p.sendSMS(phone, codeStr); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) newCode(userID string) (code *Code, codeStr string) {
	createdAt := p.Clock.NowUTC()
	codeStr = GenerateCode()
	expireAt := createdAt.Add(p.Config.ResetCodeExpiry.Duration())
	code = &Code{
		CodeHash:  HashCode(codeStr),
		UserID:    userID,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		Consumed:  false,
	}
	return
}

func (p *Provider) sendEmail(email string, code string) error {
	u := p.URLs.ResetPasswordURL(code)

	data := TemplateData{
		Email: email,
		Code:  code,
		Link:  u.String(),
	}

	msg, err := p.Translation.EmailMessageData(messageForgotPassword, data)
	if err != nil {
		return err
	}

	err = p.RateLimiter.TakeToken(mail.RateLimitBucket(email))
	if err != nil {
		return err
	}

	p.TaskQueue.Enqueue(&tasks.SendMessagesParam{
		EmailMessages: []mail.SendOptions{{
			Sender:    msg.Sender,
			ReplyTo:   msg.ReplyTo,
			Subject:   msg.Subject,
			Recipient: email,
			TextBody:  msg.TextBody,
			HTMLBody:  msg.HTMLBody,
		}},
	})

	return nil
}

func (p *Provider) sendSMS(phone string, code string) (err error) {
	u := p.URLs.ResetPasswordURL(code)

	data := TemplateData{
		Code: code,
		Link: u.String(),
	}

	msg, err := p.Translation.SMSMessageData(messageForgotPassword, data)
	if err != nil {
		return err
	}

	err = p.RateLimiter.TakeToken(sms.RateLimitBucket(phone))
	if err != nil {
		return err
	}

	p.TaskQueue.Enqueue(&tasks.SendMessagesParam{
		SMSMessages: []sms.SendOptions{{
			Sender: msg.Sender,
			To:     phone,
			Body:   msg.Body,
		}},
	})

	return
}

// ResetPassword consumes code and reset password to newPassword.
// If the code is invalid, ErrInvalidCode is returned.
// If the code is found but expired, ErrExpiredCode is returned.
// if the code is found but used, ErrUsedCode is returned.
// Otherwise, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (p *Provider) ResetPasswordByCode(codeStr string, newPassword string) (oldInfo *authenticator.Info, newInfo *authenticator.Info, err error) {
	err = p.RateLimiter.TakeToken(VerifyIPRateLimitBucket(httputil.GetIP(p.Request, bool(p.TrustProxy))))
	if err != nil {
		return
	}

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

	return p.ResetPassword(code.UserID, newPassword)
}

func (p *Provider) ResetPassword(userID string, newPassword string) (oldInfo *authenticator.Info, newInfo *authenticator.Info, err error) {
	// First see if the user has password authenticator.
	ais, err := p.Authenticators.List(
		userID,
		authenticator.KeepType(authn.AuthenticatorTypePassword),
		authenticator.KeepKind(authenticator.KindPrimary),
	)
	if err != nil {
		return
	}

	// Ensure user has password authenticator
	if len(ais) == 0 {
		err = ErrNoPassword
		return
	}

	if len(ais) == 1 {
		p.Logger.Debugf("resetting password")
		// The user has 1 password. Reset it.
		var changed bool
		var ai *authenticator.Info
		changed, ai, err = p.Authenticators.WithSecret(ais[0], newPassword)
		if err != nil {
			return
		}
		if changed {
			oldInfo = ais[0]
			newInfo = ai
		}
	} else {
		// Otherwise the user has two passwords :(
		err = fmt.Errorf("forgotpassword: detected user %s having more than 1 password", userID)
		return
	}

	return
}

func (p *Provider) HashCode(code string) string {
	return HashCode(code)
}

func (p *Provider) AfterResetPasswordByCode(codeHash string) (err error) {
	code, err := p.Store.Get(codeHash)
	if err != nil {
		return
	}

	code.Consumed = true
	err = p.Store.Update(code)
	if err != nil {
		return
	}

	return
}
