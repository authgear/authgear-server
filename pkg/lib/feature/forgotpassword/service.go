package forgotpassword

import (
	"errors"
	"strings"
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("forgot-password")}
}

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
}

type IdentityService interface {
	ListByClaim(name string, value string) ([]*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(info *authenticator.Info, options *service.UpdatePasswordOptions) (bool, *authenticator.Info, error)
	Update(info *authenticator.Info) error
	Create(authenticatorInfo *authenticator.Info, markVerified bool) error
	Delete(info *authenticator.Info) error
}

type OTPCodeService interface {
	GenerateOTP(kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	InspectState(kind otp.Kind, target string) (*otp.State, error)
	LookupCode(purpose otp.Purpose, code string) (target string, err error)
	ConsumeCode(purpose otp.Purpose, target string) error
}

type OTPSender interface {
	Prepare(channel model.AuthenticatorOOBChannel, target string, form otp.Form, typ translation.MessageType) (*otp.PreparedMessage, error)
	Send(msg *otp.PreparedMessage, opts otp.SendOptions) error
}

type Service struct {
	Logger        Logger
	Config        *config.AppConfig
	FeatureConfig *config.FeatureConfig

	Identities     IdentityService
	Authenticators AuthenticatorService
	OTPCodes       OTPCodeService
	OTPSender      OTPSender
	PasswordSender Sender

	Events EventService
}

type CodeKind string

const (
	CodeKindUnknown   CodeKind = ""
	CodeKindLink      CodeKind = "CodeKindLink"
	CodeKindShortCode CodeKind = "CodeKindShortCode"
)

type CodeChannel string

const (
	CodeChannelUnknown  CodeChannel = ""
	CodeChannelEmail    CodeChannel = "email"
	CodeChannelWhatsapp CodeChannel = "whatsapp"
	CodeChannelSMS      CodeChannel = "sms"
)

type CodeOptions struct {
	AuthenticationFlowType        string
	AuthenticationFlowName        string
	AuthenticationFlowJSONPointer jsonpointer.T
	Kind                          CodeKind
	Channel                       CodeChannel
	IsAdminAPIResetPassword       bool
}

// SendCode uses loginID to look up Email Login IDs and Phone Number Login IDs.
// For each looked up login ID, a code is generated and delivered asynchronously.
func (s *Service) SendCode(loginID string, options *CodeOptions) error {
	if options == nil {
		options = &CodeOptions{}
	}
	if !*s.Config.ForgotPassword.Enabled {
		return ErrFeatureDisabled
	}

	emailIdentities, err := s.Identities.ListByClaim(string(model.ClaimEmail), loginID)
	if err != nil {
		return err
	}
	phoneIdentities, err := s.Identities.ListByClaim(string(model.ClaimPhoneNumber), loginID)
	if err != nil {
		return err
	}

	allIdentities := append(emailIdentities, phoneIdentities...)
	if len(allIdentities) == 0 {
		// We still generate a dummy otp so that rate limits and cooldowns are still applied
		err = s.generateDummyOTP(loginID, options)
		if err != nil {
			return err
		}
		return ErrUserNotFound
	}

	for _, info := range emailIdentities {
		if !info.Type.SupportsPassword() {
			continue
		}

		standardClaims := info.IdentityAwareStandardClaims()
		email := standardClaims[model.ClaimEmail]
		if err := s.sendEmail(email, info.UserID, options); err != nil {
			return err
		}
	}

	for _, info := range phoneIdentities {
		if !info.Type.SupportsPassword() {
			continue
		}

		standardClaims := info.IdentityAwareStandardClaims()
		phone := standardClaims[model.ClaimPhoneNumber]
		if err := s.sendToPhone(phone, info.UserID, options); err != nil {
			return err
		}
	}

	return nil
}

// List out all primary password the user has.
func (s *Service) getPrimaryPasswordList(userID string) ([]*authenticator.Info, error) {
	return s.Authenticators.List(
		userID,
		authenticator.KeepType(model.AuthenticatorTypePassword),
		authenticator.KeepKind(authenticator.KindPrimary),
	)
}

func (s *Service) getForgotPasswordOTP(channel model.AuthenticatorOOBChannel, codeKind CodeKind) (otp.Kind, otp.Form) {
	switch codeKind {
	case CodeKindShortCode:
		return otp.KindForgotPasswordOTP(s.Config, channel), otp.FormCode
	case CodeKindLink:
		fallthrough
	default:
		return otp.KindForgotPasswordLink(s.Config, channel), otp.FormLink
	}
}

func (s *Service) generateDummyOTP(target string, options *CodeOptions) error {
	// Generate dummy otp for rate limiting
	otpKind, otpForm := s.getForgotPasswordOTP(s.getChannel(target, options.Channel), options.Kind)
	_, err := s.OTPCodes.GenerateOTP(
		otpKind,
		target,
		otpForm,
		&otp.GenerateOptions{
			UserID:                        "",
			AuthenticationFlowType:        options.AuthenticationFlowType,
			AuthenticationFlowName:        options.AuthenticationFlowName,
			AuthenticationFlowJSONPointer: options.AuthenticationFlowJSONPointer,
		})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) sendEmail(email string, userID string, options *CodeOptions) error {
	ais, err := s.getPrimaryPasswordList(userID)
	if err != nil {
		return err
	}
	ctx := otp.AdditionalContext{HasPassword: len(ais) > 0}

	otpKind, otpForm := s.getForgotPasswordOTP(model.AuthenticatorOOBChannelEmail, options.Kind)

	msg, err := s.OTPSender.Prepare(
		model.AuthenticatorOOBChannelEmail,
		email,
		otpForm,
		translation.MessageTypeForgotPassword,
	)
	if err != nil {
		return err
	}
	defer msg.Close()

	code, err := s.OTPCodes.GenerateOTP(
		otpKind,
		email,
		otpForm,
		&otp.GenerateOptions{
			UserID:                        userID,
			AuthenticationFlowType:        options.AuthenticationFlowType,
			AuthenticationFlowName:        options.AuthenticationFlowName,
			AuthenticationFlowJSONPointer: options.AuthenticationFlowJSONPointer,
		})
	if err != nil {
		return err
	}

	err = s.OTPSender.Send(msg, otp.SendOptions{
		OTP:                     code,
		AdditionalContext:       &ctx,
		IsAdminAPIResetPassword: options.IsAdminAPIResetPassword,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) sendToPhone(phone string, userID string, options *CodeOptions) (err error) {
	ais, err := s.getPrimaryPasswordList(userID)
	if err != nil {
		return err
	}
	ctx := otp.AdditionalContext{HasPassword: len(ais) > 0}

	if s.FeatureConfig.Identity.LoginID.Types.Phone.Disabled {
		return feature.ErrFeatureDisabledSendingSMS
	}

	otpChannel := s.getChannel(phone, options.Channel)
	var msgType translation.MessageType

	switch options.Channel {
	case CodeChannelWhatsapp:
		msgType = translation.MessageTypeWhatsappCode
	case CodeChannelSMS:
		fallthrough
	default:
		msgType = translation.MessageTypeForgotPassword
	}

	otpKind, otpForm := s.getForgotPasswordOTP(otpChannel, options.Kind)

	msg, err := s.OTPSender.Prepare(
		otpChannel,
		phone,
		otpForm,
		msgType,
	)
	if err != nil {
		return err
	}
	defer msg.Close()

	code, err := s.OTPCodes.GenerateOTP(
		otpKind,
		phone,
		otpForm,
		&otp.GenerateOptions{
			UserID:                        userID,
			AuthenticationFlowType:        options.AuthenticationFlowType,
			AuthenticationFlowName:        options.AuthenticationFlowName,
			AuthenticationFlowJSONPointer: options.AuthenticationFlowJSONPointer,
		})
	if err != nil {
		return err
	}

	err = s.OTPSender.Send(msg, otp.SendOptions{
		OTP:               code,
		AdditionalContext: &ctx,
	})
	if err != nil {
		return err
	}

	return
}

func (s *Service) getChannel(target string, codeChannel CodeChannel) model.AuthenticatorOOBChannel {
	switch codeChannel {
	case CodeChannelEmail:
		return model.AuthenticatorOOBChannelEmail
	case CodeChannelWhatsapp:
		return model.AuthenticatorOOBChannelWhatsapp
	case CodeChannelSMS:
		return model.AuthenticatorOOBChannelSMS
	default:
		// The channel is unknown, guess a value according to target
		isEmail := strings.ContainsRune(target, '@')

		var channel model.AuthenticatorOOBChannel
		if isEmail {
			channel = model.AuthenticatorOOBChannelEmail
		} else {
			switch codeChannel {
			case CodeChannelWhatsapp:
				channel = model.AuthenticatorOOBChannelWhatsapp
			case CodeChannelSMS:
				fallthrough
			default:
				channel = model.AuthenticatorOOBChannelSMS
			}
		}
		return channel
	}
}

func (s *Service) doVerifyCodeWithTarget(target string, code string, codeChannel CodeChannel, codeKind CodeKind) (state *otp.State, err error) {
	channel := s.getChannel(target, codeChannel)

	kind, otpForm := s.getForgotPasswordOTP(channel, codeKind)

	defer func() {
		if err != nil {
			err = errorutil.WithDetails(err, errorutil.Details{
				"otp_form": apierrors.APIErrorDetail.Value(otpForm),
			})
		}
	}()

	// We do not use s.InspectState() because it does not treat dummy code as invalid.
	//
	// If test mode is disabled, the dummy code is not actually sent.
	// So most of time, we will not go thought the code path of state.UserID == "".
	//
	// If test mode is enabled, the dummy code is not actually sent but a magic code can be used instead.
	// The user ID associated with the magic code is empty, violating the assumption of this package.
	state, err = s.OTPCodes.InspectState(kind, target)
	if errors.Is(err, otp.ErrConsumedCode) {
		err = ErrUsedCode
		return
	} else if apierrors.IsKind(err, otp.InvalidOTPCode) {
		err = ErrInvalidCode
		return
	} else if err != nil {
		return
	} else if state.UserID == "" {
		err = ErrInvalidCode
		return
	}

	err = s.OTPCodes.VerifyOTP(kind, target, code, &otp.VerifyOptions{
		UserID:      state.UserID,
		SkipConsume: true,
	})
	if errors.Is(err, otp.ErrConsumedCode) {
		err = ErrUsedCode
		return
	} else if apierrors.IsKind(err, otp.InvalidOTPCode) {
		err = ErrInvalidCode
		return
	} else if err != nil {
		return
	}
	return
}

func (s *Service) doVerifyCode(code string) (target string, state *otp.State, err error) {
	target, err = s.OTPCodes.LookupCode(otp.PurposeForgotPassword, code)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		err = ErrInvalidCode
		return
	} else if err != nil {
		return
	}

	state, err = s.doVerifyCodeWithTarget(target, code, CodeChannelUnknown, CodeKindUnknown)
	if err != nil {
		return
	}
	return target, state, err
}

func (s *Service) VerifyCodeWithTarget(target string, code string, codeChannel CodeChannel, kind CodeKind) (state *otp.State, err error) {
	if !*s.Config.ForgotPassword.Enabled {
		return nil, ErrFeatureDisabled
	}
	state, err = s.doVerifyCodeWithTarget(target, code, codeChannel, kind)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *Service) VerifyCode(code string) (state *otp.State, err error) {
	if !*s.Config.ForgotPassword.Enabled {
		return nil, ErrFeatureDisabled
	}

	_, state, err = s.doVerifyCode(code)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *Service) CodeLength(target string, channel CodeChannel, kind CodeKind) int {
	_, form := s.getForgotPasswordOTP(s.getChannel(target, channel), kind)
	return form.CodeLength()
}

func (s *Service) IsRateLimitError(err error, target string, channel CodeChannel, kind CodeKind) bool {
	otpKind, _ := s.getForgotPasswordOTP(s.getChannel(target, channel), kind)
	return ratelimit.IsRateLimitErrorWithBucketName(err, otpKind.RateLimitTriggerCooldown(target).Name)
}

// InspectState is for external use. It DOES NOT report dummy code as invalid.
func (s *Service) InspectState(target string, channel CodeChannel, kind CodeKind) (*otp.State, error) {
	otpKind, _ := s.getForgotPasswordOTP(s.getChannel(target, channel), kind)
	return s.OTPCodes.InspectState(otpKind, target)
}

// ResetPasswordByEndUser consumes code and reset password to newPassword.
// If the code is valid, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (s *Service) ResetPasswordByEndUser(code string, newPassword string) error {
	if !*s.Config.ForgotPassword.Enabled {
		return ErrFeatureDisabled
	}

	target, state, err := s.doVerifyCode(code)
	if err != nil {
		return err
	}

	err = s.resetPassword(target, state, newPassword, CodeChannelUnknown)
	if err != nil {
		return err
	}
	return nil
}

// ResetPasswordWithTarget is same as ResetPassword, except target is passed by caller.
func (s *Service) ResetPasswordWithTarget(target string, code string, newPassword string, channel CodeChannel, kind CodeKind) error {
	if !*s.Config.ForgotPassword.Enabled {
		return ErrFeatureDisabled
	}

	state, err := s.doVerifyCodeWithTarget(target, code, channel, kind)
	if err != nil {
		return err
	}

	err = s.resetPassword(target, state, newPassword, channel)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) resetPassword(target string, otpState *otp.State, newPassword string, channel CodeChannel) error {
	err := s.setPassword(&SetPasswordOptions{
		UserID:         otpState.UserID,
		PlainPassword:  newPassword,
		SetExpireAfter: true,
	})
	if err != nil {
		return err
	}

	err = s.OTPCodes.ConsumeCode(otp.PurposeForgotPassword, target)
	if err != nil {
		return err
	}

	// err = s.Events.DispatchEventOnCommit(ctx, &nonblocking.PasswordPrimaryResetEventPayload{
	// 	UserRef: model.UserRef{
	// 		Meta: model.Meta{
	// 			ID: otpState.UserID,
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

type SetPasswordOptions struct {
	UserID         string
	PlainPassword  string
	SetExpireAfter bool
	ExpireAfter    *time.Time
	SendPassword   bool
}

// SetPassword ensures the user identified by userID has the specified password.
// It perform necessary mutation to make this happens.
func (s *Service) setPassword(options *SetPasswordOptions) (err error) {
	ais, err := s.getPrimaryPasswordList(options.UserID)
	if err != nil {
		return
	}

	// The normal case: the user has 1 primary password
	if len(ais) == 1 {
		s.Logger.Debugf("resetting password")
		// The user has 1 password. Reset it.
		var changed bool
		var ai *authenticator.Info
		changed, ai, err = s.Authenticators.UpdatePassword(ais[0], &service.UpdatePasswordOptions{
			SetPassword:    true,
			PlainPassword:  options.PlainPassword,
			SetExpireAfter: options.SetExpireAfter,
			ExpireAfter:    options.ExpireAfter,
		})
		if err != nil {
			return
		}
		if changed {
			err = s.Authenticators.Update(ai)
			if err != nil {
				return
			}

			if options.SendPassword {
				err = s.PasswordSender.Send(options.UserID, options.PlainPassword, translation.MessageTypeSendPasswordToExistingUser)
				if err != nil {
					return
				}
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

			err = s.Authenticators.Delete(ai)
			if err != nil {
				return
			}
		}

		var newInfo *authenticator.Info
		newInfo, err = s.Authenticators.New(&authenticator.Spec{
			Type:      model.AuthenticatorTypePassword,
			Kind:      authenticator.KindPrimary,
			UserID:    options.UserID,
			IsDefault: isDefault,
			Password: &authenticator.PasswordSpec{
				PlainPassword: options.PlainPassword,
			},
		})
		if err != nil {
			return
		}

		err = s.Authenticators.Create(newInfo, true)
		if err != nil {
			return
		}

		if options.SendPassword {
			err = s.PasswordSender.Send(options.UserID, options.PlainPassword, translation.MessageTypeSendPasswordToNewUser)
			if err != nil {
				return
			}
		}
	}

	return
}

func (s *Service) ChangePasswordByAdmin(options *SetPasswordOptions) error {
	return s.setPassword(options)
}
