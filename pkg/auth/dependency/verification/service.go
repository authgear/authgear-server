package verification

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/log"
	"github.com/authgear/authgear-server/pkg/otp"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package verification

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
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
	Clock            clock.Clock
	Identities       IdentityService
	Authenticators   AuthenticatorService
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

func (s *Service) isIdentityVerified(i *identity.Info, ais []*authenticator.Info) bool {
	if !s.IsIdentityVerifiable(i) {
		return false
	}

	switch i.Type {
	case authn.IdentityTypeLoginID:
		matched := false
		filter := authenticator.KeepMatchingAuthenticatorOfIdentity(i)
		for _, a := range ais {
			if filter.Keep(a) {
				matched = true
				break
			}
		}
		return matched
	case authn.IdentityTypeOAuth:
		return true
	default:
		return false
	}
}

func (s *Service) IsIdentityVerified(i *identity.Info) (bool, error) {
	if !s.IsIdentityVerifiable(i) {
		return false, nil
	}

	authenticators, err := s.Authenticators.List(i.UserID)
	if err != nil {
		return false, err
	}

	return s.isIdentityVerified(i, authenticators), nil
}

func (s *Service) IsUserVerified(userID string) (bool, error) {
	is, err := s.Identities.ListByUser(userID)
	if err != nil {
		return false, err
	}

	as, err := s.Authenticators.List(userID)
	if err != nil {
		return false, err
	}

	return s.IsVerified(is, as), nil
}

func (s *Service) IsVerified(identities []*identity.Info, authenticators []*authenticator.Info) bool {
	numVerifiable := 0
	numVerified := 0
	for _, i := range identities {
		if !s.IsIdentityVerifiable(i) {
			continue
		}

		numVerifiable++

		if s.isIdentityVerified(i, authenticators) {
			numVerified++
		}
	}

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
		ExpireAt:   s.Clock.NowUTC().Add(s.Config.CodeExpiry.Duration()),
	}

	err := s.Store.Create(codeModel)
	if err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (s *Service) VerifyCode(id string, code string) error {
	codeModel, err := s.Store.Get(id)
	if errors.Is(err, ErrCodeNotFound) {
		return ErrInvalidVerificationCode
	} else if err != nil {
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
