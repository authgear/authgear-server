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
	New(spec *authenticator.Spec, secret string) (*authenticator.Info, error)
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

func (s *Service) isLoginIDKeyRequired(key string) bool {
	for _, c := range s.LoginID.Keys {
		if c.Key == key {
			return *c.Verification.Required
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

func (s *Service) getVerificationStatus(i *identity.Info, ais []*authenticator.Info) Status {
	if !s.IsIdentityVerifiable(i) {
		return StatusDisabled
	}

	var isVerified bool
	var isRequired bool
	switch i.Type {
	case authn.IdentityTypeLoginID:
		isVerified = false
		filter := authenticator.KeepMatchingAuthenticatorOfIdentity(i)
		for _, a := range ais {
			if filter.Keep(a) {
				isVerified = true
				break
			}
		}

		loginIDKey := i.Claims[identity.IdentityClaimLoginIDKey].(string)
		isRequired = s.isLoginIDKeyRequired(loginIDKey)
	case authn.IdentityTypeOAuth:
		isVerified = true
		isRequired = false
	default:
		isVerified = false
		isRequired = false
	}

	switch {
	case isVerified:
		return StatusVerified
	case isRequired:
		return StatusRequired
	default:
		return StatusPending
	}
}

func (s *Service) GetVerificationStatus(i *identity.Info) (Status, error) {
	if !s.IsIdentityVerifiable(i) {
		return StatusDisabled, nil
	}

	authenticators, err := s.Authenticators.List(i.UserID)
	if err != nil {
		return "", err
	}

	return s.getVerificationStatus(i, authenticators), nil
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
		status := s.getVerificationStatus(i, authenticators)
		switch status {
		case StatusVerified:
			numVerifiable++
			numVerified++
		case StatusPending, StatusRequired:
			numVerifiable++
		case StatusDisabled:
			break
		default:
			panic("verification: unknown status:" + status)
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
		ID:          id,
		UserID:      info.UserID,
		IdentityID:  info.ID,
		LoginIDType: string(loginIDType),
		LoginID:     info.Claims[identity.IdentityClaimLoginIDValue].(string),
		Code:        code,
		ExpireAt:    s.Clock.NowUTC().Add(s.Config.CodeExpiry.Duration()),
	}

	err := s.Store.Create(codeModel)
	if err != nil {
		return nil, err
	}

	return codeModel, nil
}

func (s *Service) GetCode(id string) (*Code, error) {
	return s.Store.Get(id)
}

func (s *Service) VerifyCode(id string, code string) (*Code, error) {
	codeModel, err := s.Store.Get(id)
	if errors.Is(err, ErrCodeNotFound) {
		return nil, ErrInvalidVerificationCode
	} else if err != nil {
		return nil, err
	}

	if !otp.ValidateOTP(code, codeModel.Code) {
		return nil, ErrInvalidVerificationCode
	}

	if err = s.Store.Delete(id); err != nil {
		s.Logger.WithError(err).Error("failed to delete code after verification")
	}

	return codeModel, nil
}

func (s *Service) NewVerificationAuthenticator(code *Code) (*authenticator.Info, error) {
	spec := &authenticator.Spec{
		UserID: code.UserID,
		Type:   authn.AuthenticatorTypeOOB,
		Props:  map[string]interface{}{},
	}
	switch config.LoginIDKeyType(code.LoginIDType) {
	case config.LoginIDKeyTypeEmail:
		spec.Props[authenticator.AuthenticatorPropOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelEmail)
		spec.Props[authenticator.AuthenticatorPropOOBOTPEmail] = code.LoginID
	case config.LoginIDKeyTypePhone:
		spec.Props[authenticator.AuthenticatorPropOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelSMS)
		spec.Props[authenticator.AuthenticatorPropOOBOTPPhone] = code.LoginID
	default:
		panic("verification: unsupported login ID type: " + code.LoginIDType)
	}

	return s.Authenticators.New(spec, "")
}

func (s *Service) SendCode(code *Code, url string) (*otp.CodeSendResult, error) {
	opts := otp.SendOptions{
		OTP:         code.Code,
		URL:         url,
		MessageType: otp.MessageTypeVerification,
	}

	var err error
	var channel string
	switch config.LoginIDKeyType(code.LoginIDType) {
	case config.LoginIDKeyTypeEmail:
		channel = string(authn.AuthenticatorOOBChannelEmail)
		err = s.OTPMessageSender.SendEmail(code.LoginID, opts, s.Config.Email.Message)
	case config.LoginIDKeyTypePhone:
		channel = string(authn.AuthenticatorOOBChannelSMS)
		err = s.OTPMessageSender.SendSMS(code.LoginID, opts, s.Config.SMS.Message)
	default:
		panic("verification: unsupported login ID type: " + code.LoginIDType)
	}
	if err != nil {
		return nil, err
	}

	return &otp.CodeSendResult{
		Channel:      channel,
		CodeLength:   len(code.Code),
		SendCooldown: SendCooldownSeconds,
	}, nil
}
