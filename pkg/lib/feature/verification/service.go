package verification

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package verification

type IdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

type AuthenticatorService interface {
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec, secret string) (*authenticator.Info, error)
}

type Store interface {
	Create(code *Code) error
	Get(id string) (*Code, error)
	Delete(id string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("verification")} }

type Service struct {
	Logger         Logger
	Config         *config.VerificationConfig
	LoginID        *config.LoginIDConfig
	Clock          clock.Clock
	Identities     IdentityService
	Authenticators AuthenticatorService
	Store          Store
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

func (s *Service) getVerificationStatus(i *identity.Info, iis []*identity.Info, ais []*authenticator.Info) Status {
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

		if email, ok := i.Claims[identity.StandardClaimEmail].(string); ok {
			for _, si := range iis {
				if si.ID == i.ID || si.Claims[identity.StandardClaimEmail] != email {
					continue
				}

				status := s.getVerificationStatus(si, iis, ais)
				if status == StatusVerified {
					isVerified = true
					break
				}
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

	identities, err := s.Identities.ListByUser(i.UserID)
	if err != nil {
		return "", err
	}

	authenticators, err := s.Authenticators.List(i.UserID)
	if err != nil {
		return "", err
	}

	return s.getVerificationStatus(i, identities, authenticators), nil
}

func (s *Service) GetVerificationStatuses(is []*identity.Info) (map[string]Status, error) {
	if len(is) == 0 {
		return nil, nil
	}

	// Assuming user ID of all identities is same
	userID := is[0].UserID
	authenticators, err := s.Authenticators.List(userID)
	if err != nil {
		return nil, err
	}
	identities, err := s.Identities.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	statuses := map[string]Status{}
	for _, i := range is {
		if i.UserID != userID {
			panic("verification: expect all user ID is same")
		}
		statuses[i.ID] = s.getVerificationStatus(i, identities, authenticators)
	}
	return statuses, nil
}

func (s *Service) IsUserVerified(identities []*identity.Info, userID string) (bool, error) {
	as, err := s.Authenticators.List(userID)
	if err != nil {
		return false, err
	}

	return s.IsVerified(identities, as), nil
}

func (s *Service) IsVerified(identities []*identity.Info, authenticators []*authenticator.Info) bool {
	numVerifiable := 0
	numVerified := 0
	for _, i := range identities {
		status := s.getVerificationStatus(i, identities, authenticators)
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
		ID:           id,
		UserID:       info.UserID,
		IdentityID:   info.ID,
		IdentityType: string(info.Type),
		LoginIDType:  string(loginIDType),
		LoginID:      info.Claims[identity.IdentityClaimLoginIDValue].(string),
		Code:         code,
		ExpireAt:     s.Clock.NowUTC().Add(s.Config.CodeExpiry.Duration()),
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
		Claims: map[string]interface{}{},
	}
	switch config.LoginIDKeyType(code.LoginIDType) {
	case config.LoginIDKeyTypeEmail:
		spec.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelEmail)
		spec.Claims[authenticator.AuthenticatorClaimOOBOTPEmail] = code.LoginID
	case config.LoginIDKeyTypePhone:
		spec.Claims[authenticator.AuthenticatorClaimOOBOTPChannelType] = string(authn.AuthenticatorOOBChannelSMS)
		spec.Claims[authenticator.AuthenticatorClaimOOBOTPPhone] = code.LoginID
	default:
		panic("verification: unsupported login ID type: " + code.LoginIDType)
	}

	return s.Authenticators.New(spec, "")
}
