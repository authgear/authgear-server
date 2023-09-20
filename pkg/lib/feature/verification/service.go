package verification

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package verification

type ClaimStore interface {
	ListByUser(userID string) ([]*Claim, error)
	ListByClaimName(userID string, claimName string) ([]*Claim, error)
	Get(userID string, claimName string, claimValue string) (*Claim, error)
	Create(claim *Claim) error
	Delete(id string) error
	DeleteAll(userID string) error
}

type Service struct {
	Config            *config.VerificationConfig
	UserProfileConfig *config.UserProfileConfig

	Clock      clock.Clock
	ClaimStore ClaimStore
}

func (s *Service) claimVerificationConfig(claimName model.ClaimName) *config.VerificationClaimConfig {
	switch claimName {
	case model.ClaimEmail:
		return s.Config.Claims.Email
	case model.ClaimPhoneNumber:
		return s.Config.Claims.PhoneNumber
	default:
		return nil
	}
}

func (s *Service) getVerificationStatus(i *identity.Info, verifiedClaims map[claim]struct{}) []ClaimStatus {
	var statuses []ClaimStatus
	standardClaims := i.IdentityAwareStandardClaims()
	for claimName, claimValue := range standardClaims {
		c := s.claimVerificationConfig(claimName)
		if c == nil {
			continue
		}

		value := claimValue

		_, verified := verifiedClaims[claim{string(claimName), value}]

		statuses = append(statuses, ClaimStatus{
			Name:                       string(claimName),
			Value:                      value,
			Verified:                   verified,
			RequiredToVerifyOnCreation: *c.Required,
			EndUserTriggerable:         *c.Enabled,
		})
	}
	return statuses
}

func (s *Service) GetIdentityVerificationStatus(i *identity.Info) ([]ClaimStatus, error) {
	claims, err := s.ClaimStore.ListByUser(i.UserID)
	if err != nil {
		return nil, err
	}

	verifiedClaims := make(map[claim]struct{})
	for _, c := range claims {
		verifiedClaims[claim{c.Name, c.Value}] = struct{}{}
	}

	return s.getVerificationStatus(i, verifiedClaims), nil
}

func (s *Service) GetVerificationStatuses(is []*identity.Info) (map[string][]ClaimStatus, error) {
	if len(is) == 0 {
		return nil, nil
	}

	// Assuming user ID of all identities is same
	userID := is[0].UserID
	claims, err := s.ClaimStore.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	verifiedClaims := make(map[claim]struct{})
	for _, c := range claims {
		verifiedClaims[claim{c.Name, c.Value}] = struct{}{}
	}

	statuses := map[string][]ClaimStatus{}
	for _, i := range is {
		if i.UserID != userID {
			panic("verification: expect all user ID is same")
		}
		statuses[i.ID] = s.getVerificationStatus(i, verifiedClaims)
	}
	return statuses, nil
}

func (s *Service) GetAuthenticatorVerificationStatus(a *authenticator.Info) (AuthenticatorStatus, error) {
	if a.Type != model.AuthenticatorTypeOOBEmail && a.Type != model.AuthenticatorTypeOOBSMS {
		panic("verification: incompatible authenticator type: " + a.Type)
	}

	var claimName string
	var claimValue string
	aClaims := a.StandardClaims()
	switch a.Type {
	case model.AuthenticatorTypeOOBEmail:
		claimName = string(model.ClaimEmail)
		claimValue = aClaims[model.ClaimEmail]
	case model.AuthenticatorTypeOOBSMS:
		claimName = string(model.ClaimPhoneNumber)
		claimValue = aClaims[model.ClaimPhoneNumber]
	}

	_, err := s.ClaimStore.Get(a.UserID, claimName, claimValue)
	if errors.Is(err, ErrClaimUnverified) {
		return AuthenticatorStatusUnverified, nil
	} else if err != nil {
		return "", err
	}

	return AuthenticatorStatusVerified, nil
}

func (s *Service) GetClaims(userID string) ([]*Claim, error) {
	return s.ClaimStore.ListByUser(userID)
}

func (s *Service) GetClaimStatus(userID string, claimName model.ClaimName, claimValue string) (*ClaimStatus, error) {
	claims, err := s.ClaimStore.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	cfg := s.claimVerificationConfig(claimName)
	if cfg == nil {
		return nil, ErrUnsupportedClaim
	}

	verified := false
	for _, claim := range claims {
		if claim.Name == string(claimName) && claim.Value == claimValue {
			verified = true
		}
	}

	return &ClaimStatus{
		Name:                       string(claimName),
		Value:                      claimValue,
		Verified:                   verified,
		RequiredToVerifyOnCreation: *cfg.Required,
		EndUserTriggerable:         *cfg.Enabled,
	}, nil
}

func (s *Service) IsUserVerified(identities []*identity.Info) (bool, error) {
	statuses, err := s.GetVerificationStatuses(identities)
	if err != nil {
		return false, err
	}

	numVerifiable := 0
	numVerified := 0
	for _, claimStatuses := range statuses {
		for _, claim := range claimStatuses {
			if claim.IsVerifiable() {
				numVerifiable++
			}
			if claim.Verified {
				numVerified++
			}
		}
	}

	switch s.Config.Criteria {
	case config.VerificationCriteriaAny:
		return numVerifiable > 0 && numVerified >= 1, nil
	case config.VerificationCriteriaAll:
		return numVerifiable > 0 && numVerified == numVerifiable, nil
	default:
		panic("verification: unknown criteria " + s.Config.Criteria)
	}
}

func (s *Service) NewVerifiedClaim(userID string, claimName string, claimValue string) *Claim {
	return &Claim{
		ID:     uuid.New(),
		UserID: userID,
		Name:   claimName,
		Value:  claimValue,
	}
}

func (s *Service) MarkClaimVerified(claim *Claim) error {
	claims, err := s.GetClaims(claim.UserID)
	if err != nil {
		return err
	}
	for _, c := range claims {
		if c.Name == claim.Name && c.Value == claim.Value {
			return nil
		}
	}
	claim.CreatedAt = s.Clock.NowUTC()
	return s.ClaimStore.Create(claim)
}

func (s *Service) DeleteClaim(claim *Claim) error {
	return s.ClaimStore.Delete(claim.ID)
}

func (s *Service) ResetVerificationStatus(userID string) error {
	return s.ClaimStore.DeleteAll(userID)
}

func (s *Service) RemoveOrphanedClaims(userID string, identities []*identity.Info, authenticators []*authenticator.Info) error {
	claims, err := s.ClaimStore.ListByUser(userID)
	if err != nil {
		return err
	}

	orphans := make(map[claim]*Claim)
	for _, c := range claims {
		orphans[claim{c.Name, c.Value}] = c
	}

	for _, i := range identities {
		if i.UserID != userID {
			panic("verification: expect all user ID is same")
		}
		for name, value := range i.IdentityAwareStandardClaims() {
			delete(orphans, claim{Name: string(name), Value: value})
		}
	}

	for _, a := range authenticators {
		if a.UserID != userID {
			panic("verification: expect all user ID is same")
		}
		for name, value := range a.StandardClaims() {
			delete(orphans, claim{Name: string(name), Value: value})
		}
	}

	for _, claim := range orphans {
		err = s.ClaimStore.Delete(claim.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
