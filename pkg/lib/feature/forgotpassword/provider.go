package forgotpassword

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature"
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
	Update(info *authenticator.Info) error
	Create(info *authenticator.Info) error
	Delete(info *authenticator.Info) error
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

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type HardSMSBucketer interface {
	Bucket() ratelimit.Bucket
}

type Provider struct {
	RemoteIP    httputil.RemoteIP
	Translation TranslationService
	Config      *config.ForgotPasswordConfig

	Store     *Store
	Clock     clock.Clock
	URLs      URLProvider
	TaskQueue task.Queue

	Logger ProviderLogger

	Identities     IdentityService
	Authenticators AuthenticatorService
	FeatureConfig  *config.FeatureConfig
	Events         EventService

	RateLimiter     RateLimiter
	HardSMSBucketer HardSMSBucketer
}

// SendCode uses loginID to look up Email Login IDs and Phone Number Login IDs.
// For each looked up login ID, a code is generated and delivered asynchronously.
func (p *Provider) SendCode(loginID string) error {
	err := p.RateLimiter.TakeToken(AntiSpamSendCodeBucket(loginID))
	if err != nil {
		return err
	}

	emailIdentities, err := p.Identities.ListByClaim(string(model.ClaimEmail), loginID)
	if err != nil {
		return err
	}
	phoneIdentities, err := p.Identities.ListByClaim(string(model.ClaimPhoneNumber), loginID)
	if err != nil {
		return err
	}

	allIdentities := append(emailIdentities, phoneIdentities...)
	if len(allIdentities) == 0 {
		return ErrUserNotFound
	}

	for _, info := range emailIdentities {
		email := info.Claims[string(model.ClaimEmail)].(string)
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
		phone := info.Claims[string(model.ClaimPhoneNumber)].(string)
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

	err = p.RateLimiter.TakeToken(mail.AntiSpamBucket(email))
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

	err = p.Events.DispatchEvent(&nonblocking.EmailSentEventPayload{
		Sender:    msg.Sender,
		Recipient: email,
		Type:      nonblocking.MessageTypeForgotPassword,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) sendSMS(phone string, code string) (err error) {
	fc := p.FeatureConfig
	if fc.Identity.LoginID.Types.Phone.Disabled {
		return feature.ErrFeatureDisabledSendingSMS
	}

	u := p.URLs.ResetPasswordURL(code)

	data := TemplateData{
		Code: code,
		Link: u.String(),
	}

	msg, err := p.Translation.SMSMessageData(messageForgotPassword, data)
	if err != nil {
		return err
	}

	err = p.RateLimiter.TakeToken(sms.AntiSpamBucket(phone))
	if err != nil {
		return err
	}

	err = p.RateLimiter.TakeToken(p.HardSMSBucketer.Bucket())
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

	err = p.Events.DispatchEvent(&nonblocking.SMSSentEventPayload{
		Sender:    msg.Sender,
		Recipient: phone,
		Type:      nonblocking.MessageTypeForgotPassword,
	})
	if err != nil {
		return err
	}

	return
}

// ResetPassword consumes code and reset password to newPassword.
// If the code is invalid, ErrInvalidCode is returned.
// If the code is found but expired, ErrExpiredCode is returned.
// if the code is found but used, ErrUsedCode is returned.
// Otherwise, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (p *Provider) ResetPasswordByCode(codeStr string, newPassword string) (err error) {
	err = p.RateLimiter.TakeToken(AntiBruteForceVerifyBucket(string(p.RemoteIP)))
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

	err = p.ResetPassword(code.UserID, newPassword)
	if err != nil {
		return
	}

	err = p.Store.MarkConsumed(codeHash)
	if err != nil {
		return
	}

	return
}

// ResetPassword ensures the user identified by userID has a password.
// It perform necessary mutation to make this happens.
func (p *Provider) ResetPassword(userID string, newPassword string) (err error) {
	// List out all primary password the user has.
	ais, err := p.Authenticators.List(
		userID,
		authenticator.KeepType(model.AuthenticatorTypePassword),
		authenticator.KeepKind(authenticator.KindPrimary),
	)
	if err != nil {
		return
	}

	// The normal case: the user has 1 primary password
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
			err = p.Authenticators.Update(ai)
			if err != nil {
				return
			}
		}
	} else {
		// The special case: the user either has no primary password or
		// more than 1 primary passwords.
		// We delete the existing primary passwords and then create a new one.
		isDefault := false
		for _, ai := range ais {
			// If one of the authenticator we are going to delete is default,
			// then the authenticator we are going to create should be default.
			if ai.IsDefault {
				isDefault = true
			}

			err = p.Authenticators.Delete(ai)
			if err != nil {
				return
			}
		}

		var newInfo *authenticator.Info
		newInfo, err = p.Authenticators.New(&authenticator.Spec{
			UserID:    userID,
			IsDefault: isDefault,
			Kind:      authenticator.KindPrimary,
			Type:      model.AuthenticatorTypePassword,
			Claims:    map[string]interface{}{},
		}, newPassword)
		if err != nil {
			return
		}

		err = p.Authenticators.Create(newInfo)
		if err != nil {
			return
		}
	}

	return
}
