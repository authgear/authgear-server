package forgotpassword

import (
	"errors"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("forgot-password")}
}

type messageContext struct {
	HasPassword bool
}

type IdentityService interface {
	ListByClaim(name string, value string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec) (*authenticator.Info, error)
	WithSpec(ai *authenticator.Info, spec *authenticator.Spec) (bool, *authenticator.Info, error)
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
	Prepare(channel model.AuthenticatorOOBChannel, target string, form otp.Form, typ otp.MessageType) (*otp.PreparedMessage, error)
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
}

// SendCode uses loginID to look up Email Login IDs and Phone Number Login IDs.
// For each looked up login ID, a code is generated and delivered asynchronously.
func (s *Service) SendCode(loginID string) error {
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
		// FIXME: login ID existence rate limit
		return ErrUserNotFound
	}

	for _, info := range emailIdentities {
		standardClaims := info.IdentityAwareStandardClaims()
		email := standardClaims[model.ClaimEmail]
		if err := s.sendEmail(email, info.UserID); err != nil {
			return err
		}
	}

	for _, info := range phoneIdentities {
		standardClaims := info.IdentityAwareStandardClaims()
		phone := standardClaims[model.ClaimPhoneNumber]
		if err := s.sendSMS(phone, info.UserID); err != nil {
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

func (s *Service) sendEmail(email string, userID string) error {
	ais, err := s.getPrimaryPasswordList(userID)
	if err != nil {
		return err
	}
	ctx := messageContext{HasPassword: len(ais) > 0}

	msg, err := s.OTPSender.Prepare(
		model.AuthenticatorOOBChannelEmail,
		email,
		otp.FormLink,
		otp.MessageTypeForgotPassword,
	)
	if err != nil {
		return err
	}
	defer msg.Close()

	code, err := s.OTPCodes.GenerateOTP(
		otp.KindForgotPassword(s.Config, model.AuthenticatorOOBChannelEmail),
		email,
		otp.FormLink,
		&otp.GenerateOptions{
			UserID: userID,
			// TODO: workflow/session ID?
		})
	if err != nil {
		return err
	}

	err = s.OTPSender.Send(msg, otp.SendOptions{
		OTP:               code,
		AdditionalContext: ctx,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) sendSMS(phone string, userID string) (err error) {
	ais, err := s.getPrimaryPasswordList(userID)
	if err != nil {
		return err
	}
	ctx := messageContext{HasPassword: len(ais) > 0}

	if s.FeatureConfig.Identity.LoginID.Types.Phone.Disabled {
		return feature.ErrFeatureDisabledSendingSMS
	}

	msg, err := s.OTPSender.Prepare(
		model.AuthenticatorOOBChannelSMS,
		phone,
		otp.FormLink,
		otp.MessageTypeForgotPassword,
	)
	if err != nil {
		return err
	}
	defer msg.Close()

	code, err := s.OTPCodes.GenerateOTP(
		otp.KindForgotPassword(s.Config, model.AuthenticatorOOBChannelSMS),
		phone,
		otp.FormLink,
		&otp.GenerateOptions{
			UserID: userID,
			// TODO: workflow/session ID?
		})
	if err != nil {
		return err
	}

	err = s.OTPSender.Send(msg, otp.SendOptions{
		OTP:               code,
		AdditionalContext: ctx,
	})
	if err != nil {
		return err
	}

	return
}

func (s *Service) doVerifyCode(code string) (target string, userID string, err error) {
	target, err = s.OTPCodes.LookupCode(otp.PurposeForgotPassword, code)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		err = ErrInvalidCode
		return
	} else if err != nil {
		return
	}

	// TODO: more robust?
	channel := model.AuthenticatorOOBChannelSMS
	if strings.ContainsRune(target, '@') {
		channel = model.AuthenticatorOOBChannelEmail
	}
	kind := otp.KindForgotPassword(s.Config, channel)

	state, err := s.OTPCodes.InspectState(kind, target)
	if errors.Is(err, otp.ErrConsumedCode) {
		err = ErrUsedCode
		return
	} else if apierrors.IsKind(err, otp.InvalidOTPCode) {
		err = ErrInvalidCode
		return
	} else if err != nil {
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

	userID = state.UserID
	return
}

func (s *Service) VerifyCode(code string) (userID string, err error) {
	if !*s.Config.ForgotPassword.Enabled {
		return "", ErrFeatureDisabled
	}

	_, userID, err = s.doVerifyCode(code)
	if err != nil {
		return "", err
	}

	return userID, nil
}

// ResetPassword consumes code and reset password to newPassword.
// If the code is valid, the password is reset to newPassword.
// newPassword is checked against the password policy so
// password policy error may also be returned.
func (s *Service) ResetPassword(code string, newPassword string) error {
	if !*s.Config.ForgotPassword.Enabled {
		return ErrFeatureDisabled
	}

	target, userID, err := s.doVerifyCode(code)
	if err != nil {
		return err
	}

	err = s.SetPassword(userID, newPassword)
	if err != nil {
		return err
	}

	err = s.OTPCodes.ConsumeCode(otp.PurposeForgotPassword, target)
	if err != nil {
		return err
	}

	return nil
}

// SetPassword ensures the user identified by userID has the specified password.
// It perform necessary mutation to make this happens.
func (s *Service) SetPassword(userID string, newPassword string) (err error) {
	ais, err := s.getPrimaryPasswordList(userID)
	if err != nil {
		return
	}

	// The normal case: the user has 1 primary password
	if len(ais) == 1 {
		s.Logger.Debugf("resetting password")
		// The user has 1 password. Reset it.
		var changed bool
		var ai *authenticator.Info
		changed, ai, err = s.Authenticators.WithSpec(ais[0], &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: newPassword,
			},
		})
		if err != nil {
			return
		}
		if changed {
			err = s.Authenticators.Update(ai)
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

			err = s.Authenticators.Delete(ai)
			if err != nil {
				return
			}
		}

		var newInfo *authenticator.Info
		newInfo, err = s.Authenticators.New(&authenticator.Spec{
			Type:      model.AuthenticatorTypePassword,
			Kind:      authenticator.KindPrimary,
			UserID:    userID,
			IsDefault: isDefault,
			Password: &authenticator.PasswordSpec{
				PlainPassword: newPassword,
			},
		})
		if err != nil {
			return
		}

		err = s.Authenticators.Create(newInfo, true)
		if err != nil {
			return
		}
	}

	return
}
