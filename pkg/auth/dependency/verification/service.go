package verification

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/log"
	"github.com/authgear/authgear-server/pkg/otp"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package verification

type IdentityProvider interface {
	ListByUser(userID string) ([]*identity.Info, error)
	RelateIdentityToAuthenticator(ii *identity.Info, as *authenticator.Spec) *authenticator.Spec
}

type AuthenticatorProvider interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type OTPMessageSender interface {
	SendEmail(email string, opts otp.SendOptions, message config.EmailMessageConfig) error
	SendSMS(phone string, opts otp.SendOptions, message config.SMSMessageConfig) error
}

type Store interface {
	Create(code *Code) error
	Get(id string) (*Code, error)
	Delete(id string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("verification")} }

type Service struct {
	Logger           Logger
	Config           *config.VerificationConfig
	LoginID          *config.LoginIDConfig
	Identities       IdentityProvider `wire:"-"`
	Authenticators   AuthenticatorProvider
	OTPMessageSender OTPMessageSender
	Store            Store
}

func (s *Service) isLoginIDKeyVerifiable(key string) bool {
	for _, c := range s.LoginID.Keys {
		if c.Key == key {
			return *c.Verification.Enabled
		}
	}
	return false
}

func (s *Service) IsIdentityVerifiable(i *identity.Info) bool {
	switch i.Type {
	case authn.IdentityTypeLoginID:
		key := i.Claims[identity.IdentityClaimLoginIDKey].(string)
		return s.isLoginIDKeyVerifiable(key)
	case authn.IdentityTypeOAuth:
		return true
	default:
		return false
	}
}

func (s *Service) IsIdentityVerified(i *identity.Info) (bool, error) {
	// FIXME: verification
	/*
		switch i.Type {
		case authn.IdentityTypeOAuth:
			return true, nil
		case authn.IdentityTypeLoginID:
			if !s.isLoginIDKeyVerifiable(i.Claims[identity.IdentityClaimLoginIDKey].(string)) {
				return false, nil
			}

			as, err := s.Authenticators.List(i.UserID, authn.AuthenticatorTypeOOB)
			if err != nil {
				return false, err
			}

			for _, a := range as {
				spec := a.ToSpec()
				if s.Identities.RelateIdentityToAuthenticator(i, &spec) != nil {
					return true, nil
				}
			}
			return false, nil
		default:
			return false, nil
		}*/
	return false, nil
}

func (s *Service) IsUserVerified(userID string) (bool, error) {
	// FIXME: verification
	return false, nil
	/*
		is, err := s.Identities.ListByUser(userID)
		if err != nil {
			return false, err
		}

		as, err := s.Authenticators.List(userID, authn.AuthenticatorTypeOOB)
		if err != nil {
			return false, err
		}

		return s.IsVerified(is, as), nil
	*/
}

func (s *Service) IsVerified(identities []*identity.Info, authenticators []*authenticator.Info) bool {
	numVerifiable := 0
	numVerified := 0
	// FIXME: verification
	/*
		for _, i := range identities {
			switch i.Type {
			case authn.IdentityTypeLoginID:
				if !s.isLoginIDKeyVerifiable(i.Claims[identity.IdentityClaimLoginIDKey].(string)) {
					continue
				}

				numVerifiable++
				for _, a := range authenticators {
					if a.Type != authn.AuthenticatorTypeOOB {
						continue
					}

					spec := a.ToSpec()
					if s.Identities.RelateIdentityToAuthenticator(i, &spec) != nil {
						numVerified++
						break
					}
				}
			case authn.IdentityTypeOAuth:
				numVerifiable++
				numVerified++
			default:
				continue
			}
		}*/

	switch s.Config.Criteria {
	case config.VerificationCriteriaAny:
		return numVerifiable > 0 && numVerified >= 1
	case config.VerificationCriteriaAll:
		return numVerifiable > 0 && numVerified == numVerifiable
	default:
		panic("verification: unknown criteria " + s.Config.Criteria)
	}
}

func (s *Service) CreateNewCode(id string, info *identity.Info) (*Code, error) {
	if info.Type != authn.IdentityTypeLoginID {
		panic("verification: expect login ID identity")
	}

	loginIDType := config.LoginIDKeyType(info.Claims[identity.IdentityClaimLoginIDType].(string))

	var code string
	switch loginIDType {
	case config.LoginIDKeyTypeEmail:
		code = otp.FormatComplex.Generate()
	case config.LoginIDKeyTypePhone:
		code = otp.FormatNumeric.Generate()
	default:
		panic("verification: unsupported login ID type: " + loginIDType)
	}

	codeModel := &Code{
		ID:         id,
		UserID:     info.UserID,
		IdentityID: info.ID,
		Code:       code,
	}

	err := s.Store.Create(codeModel)
	if err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (s *Service) VerifyCode(id string, code string) error {
	codeModel, err := s.Store.Get(id)
	if err != nil {
		return err
	}

	if !otp.ValidateOTP(code, codeModel.Code) {
		return ErrInvalidVerificationCode
	}

	if err = s.Store.Delete(id); err != nil {
		s.Logger.WithError(err).Error("failed to delete code after verification")
	}

	return nil
}

func (s *Service) SendCode(
	loginIDType config.LoginIDKeyType,
	loginIDValue string,
	code *Code,
	url string,
) error {
	opts := otp.SendOptions{
		OTP:         code.Code,
		URL:         url,
		MessageType: otp.MessageTypeVerification,
	}

	switch loginIDType {
	case config.LoginIDKeyTypeEmail:
		return s.OTPMessageSender.SendEmail(loginIDValue, opts, s.Config.Email.Message)
	case config.LoginIDKeyTypePhone:
		return s.OTPMessageSender.SendSMS(loginIDValue, opts, s.Config.SMS.Message)
	default:
		panic("verification: unsupported login ID type: " + loginIDType)
	}
}
