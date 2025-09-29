package forgotpassword

import (
	"context"
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
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var ForgotPasswordLogger = slogutil.NewLogger("forgot-password")

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type IdentityService interface {
	ListByClaim(ctx context.Context, name string, value string) ([]*identity.Info, error)
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	New(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error)
	UpdatePassword(ctx context.Context, info *authenticator.Info, options *service.UpdatePasswordOptions) (bool, *authenticator.Info, error)

	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	Update(ctx context.Context, info *authenticator.Info) error
	Create(ctx context.Context, authenticatorInfo *authenticator.Info, markVerified bool) error
	Delete(ctx context.Context, info *authenticator.Info) error
}

type OTPCodeService interface {
	GenerateOTP(ctx context.Context, kind otp.Kind, target string, form otp.Form, opt *otp.GenerateOptions) (string, error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	InspectState(ctx context.Context, kind otp.Kind, target string) (*otp.State, error)
	LookupCode(ctx context.Context, purpose otp.Purpose, code string) (target string, err error)
	ConsumeCode(ctx context.Context, purpose otp.Purpose, target string) error
}

type OTPSender interface {
	Send(ctx context.Context, opts otp.SendOptions) error
}

type Service struct {
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
func (s *Service) SendCode(ctx context.Context, loginID string, options *CodeOptions) error {
	if options == nil {
		options = &CodeOptions{}
	}
	if !*s.Config.ForgotPassword.Enabled {
		return ErrFeatureDisabled
	}

	emailIdentities, err := s.Identities.ListByClaim(ctx, string(model.ClaimEmail), loginID)
	if err != nil {
		return err
	}
	phoneIdentities, err := s.Identities.ListByClaim(ctx, string(model.ClaimPhoneNumber), loginID)
	if err != nil {
		return err
	}

	allIdentities := append(emailIdentities, phoneIdentities...)
	if len(allIdentities) == 0 {
		// We still generate a dummy otp so that rate limits and cooldowns are still applied
		err = s.generateDummyOTP(ctx, loginID, options)
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
		if err := s.sendEmail(ctx, email, info.UserID, options); err != nil {
			return err
		}
	}

	for _, info := range phoneIdentities {
		if !info.Type.SupportsPassword() {
			continue
		}

		standardClaims := info.IdentityAwareStandardClaims()
		phone := standardClaims[model.ClaimPhoneNumber]
		if err := s.sendToPhone(ctx, phone, info.UserID, options); err != nil {
			return err
		}
	}

	return nil
}

// List out all primary password the user has.
func (s *Service) getPrimaryPasswordList(ctx context.Context, userID string) ([]*authenticator.Info, error) {
	return s.Authenticators.List(
		ctx,
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

func (s *Service) generateDummyOTP(ctx context.Context, target string, options *CodeOptions) error {
	// Generate dummy otp for rate limiting
	otpKind, otpForm := s.getForgotPasswordOTP(s.getChannel(target, options.Channel), options.Kind)
	_, err := s.OTPCodes.GenerateOTP(
		ctx,
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

func (s *Service) sendEmail(ctx context.Context, email string, userID string, options *CodeOptions) error {
	ais, err := s.getPrimaryPasswordList(ctx, userID)
	if err != nil {
		return err
	}
	otpCtx := otp.AdditionalContext{HasPassword: len(ais) > 0}

	otpKind, otpForm := s.getForgotPasswordOTP(model.AuthenticatorOOBChannelEmail, options.Kind)

	code, err := s.OTPCodes.GenerateOTP(
		ctx,
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

	err = s.OTPSender.Send(
		ctx,
		otp.SendOptions{
			Channel:                 model.AuthenticatorOOBChannelEmail,
			Target:                  email,
			Form:                    otpForm,
			Type:                    translation.MessageTypeForgotPassword,
			Kind:                    otpKind,
			OTP:                     code,
			AdditionalContext:       &otpCtx,
			IsAdminAPIResetPassword: options.IsAdminAPIResetPassword,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) sendToPhone(ctx context.Context, phone string, userID string, options *CodeOptions) (err error) {
	ais, err := s.getPrimaryPasswordList(ctx, userID)
	if err != nil {
		return err
	}
	otpCtx := otp.AdditionalContext{HasPassword: len(ais) > 0}

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

	code, err := s.OTPCodes.GenerateOTP(
		ctx,
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

	err = s.OTPSender.Send(
		ctx,
		otp.SendOptions{
			Channel:           otpChannel,
			Target:            phone,
			Form:              otpForm,
			Type:              msgType,
			Kind:              otpKind,
			OTP:               code,
			AdditionalContext: &otpCtx,
		},
	)
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

func (s *Service) doVerifyCodeWithTarget(ctx context.Context, target string, code string, codeChannel CodeChannel, codeKind CodeKind) (state *otp.State, err error) {
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
	state, err = s.OTPCodes.InspectState(ctx, kind, target)
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

	err = s.OTPCodes.VerifyOTP(ctx, kind, target, code, &otp.VerifyOptions{
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

func (s *Service) doVerifyCode(ctx context.Context, code string) (target string, state *otp.State, err error) {
	target, err = s.OTPCodes.LookupCode(ctx, otp.PurposeForgotPassword, code)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		err = ErrInvalidCode
		return
	} else if err != nil {
		return
	}

	state, err = s.doVerifyCodeWithTarget(ctx, target, code, CodeChannelUnknown, CodeKindUnknown)
	if err != nil {
		return
	}
	return target, state, err
}

func (s *Service) VerifyCodeWithTarget(ctx context.Context, target string, code string, codeChannel CodeChannel, kind CodeKind) (state *otp.State, err error) {
	if !*s.Config.ForgotPassword.Enabled {
		return nil, ErrFeatureDisabled
	}
	state, err = s.doVerifyCodeWithTarget(ctx, target, code, codeChannel, kind)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *Service) VerifyCode(ctx context.Context, code string) (state *otp.State, err error) {
	if !*s.Config.ForgotPassword.Enabled {
		return nil, ErrFeatureDisabled
	}

	_, state, err = s.doVerifyCode(ctx, code)
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
func (s *Service) InspectState(ctx context.Context, target string, channel CodeChannel, kind CodeKind) (*otp.State, error) {
	otpKind, _ := s.getForgotPasswordOTP(s.getChannel(target, channel), kind)
	return s.OTPCodes.InspectState(ctx, otpKind, target)
}

// ResetPasswordByEndUser consumes code and reset password to newPassword.
// If the code is valid, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (s *Service) ResetPasswordByEndUser(ctx context.Context, code string, newPassword string) error {
	if !*s.Config.ForgotPassword.Enabled {
		return ErrFeatureDisabled
	}

	target, state, err := s.doVerifyCode(ctx, code)
	if err != nil {
		return err
	}

	err = s.resetPassword(ctx, target, state, newPassword, CodeChannelUnknown)
	if err != nil {
		return err
	}
	return nil
}

// ResetPasswordWithTarget is same as ResetPassword, except target is passed by caller.
func (s *Service) ResetPasswordWithTarget(ctx context.Context, target string, code string, newPassword string, channel CodeChannel, kind CodeKind) error {
	if !*s.Config.ForgotPassword.Enabled {
		return ErrFeatureDisabled
	}

	state, err := s.doVerifyCodeWithTarget(ctx, target, code, channel, kind)
	if err != nil {
		return err
	}

	err = s.resetPassword(ctx, target, state, newPassword, channel)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) resetPassword(ctx context.Context, target string, otpState *otp.State, newPassword string, channel CodeChannel) error {
	err := s.setPassword(ctx, &SetPasswordOptions{
		UserID:         otpState.UserID,
		PlainPassword:  newPassword,
		SetExpireAfter: true,
	})
	if err != nil {
		return err
	}

	err = s.OTPCodes.ConsumeCode(ctx, otp.PurposeForgotPassword, target)
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
func (s *Service) setPassword(ctx context.Context, options *SetPasswordOptions) (err error) {
	logger := ForgotPasswordLogger.GetLogger(ctx)

	ais, err := s.getPrimaryPasswordList(ctx, options.UserID)
	if err != nil {
		return
	}

	// The normal case: the user has 1 primary password
	if len(ais) == 1 {
		logger.Debug(ctx, "resetting password")
		// The user has 1 password. Reset it.
		var changed bool
		var ai *authenticator.Info
		changed, ai, err = s.Authenticators.UpdatePassword(ctx, ais[0], &service.UpdatePasswordOptions{
			SetPassword:    true,
			PlainPassword:  options.PlainPassword,
			SetExpireAfter: options.SetExpireAfter,
			ExpireAfter:    options.ExpireAfter,
		})
		if err != nil {
			return
		}
		if changed {
			err = s.Authenticators.Update(ctx, ai)
			if err != nil {
				return
			}

			if options.SendPassword {
				err = s.PasswordSender.Send(ctx, options.UserID, options.PlainPassword, translation.MessageTypeSendPasswordToExistingUser)
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

			err = s.Authenticators.Delete(ctx, ai)
			if err != nil {
				return
			}
		}

		var newInfo *authenticator.Info
		newInfo, err = s.Authenticators.New(ctx, &authenticator.Spec{
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

		err = s.Authenticators.Create(ctx, newInfo, true)
		if err != nil {
			return
		}

		if options.SendPassword {
			err = s.PasswordSender.Send(ctx, options.UserID, options.PlainPassword, translation.MessageTypeSendPasswordToNewUser)
			if err != nil {
				return
			}
		}
	}

	return
}

func (s *Service) ChangePasswordByAdmin(ctx context.Context, options *SetPasswordOptions) error {
	return s.setPassword(ctx, options)
}
