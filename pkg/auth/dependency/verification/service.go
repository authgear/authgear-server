package verification

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package verification

type IdentityProvider interface {
	ListByUser(userID string) ([]*identity.Info, error)
	RelateIdentityToAuthenticator(is identity.Spec, as *authenticator.Spec) *authenticator.Spec
}

type AuthenticatorProvider interface {
	List(userID string, typ authn.AuthenticatorType) ([]*authenticator.Info, error)
}

type Service struct {
	Config         *config.VerificationConfig
	LoginID        *config.LoginIDConfig
	Identities     IdentityProvider
	Authenticators AuthenticatorProvider
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
			if s.Identities.RelateIdentityToAuthenticator(i.ToSpec(), &spec) != nil {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

func (s *Service) IsUserVerified(userID string) (bool, error) {
	is, err := s.Identities.ListByUser(userID)
	if err != nil {
		return false, err
	}

	as, err := s.Authenticators.List(userID, authn.AuthenticatorTypeOOB)
	if err != nil {
		return false, err
	}

	return s.IsVerified(is, as), nil
}

func (s *Service) IsVerified(identities []*identity.Info, authenticators []*authenticator.Info) bool {
	numVerifiable := 0
	numVerified := 0
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
				if s.Identities.RelateIdentityToAuthenticator(i.ToSpec(), &spec) != nil {
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
