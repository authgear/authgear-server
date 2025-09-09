package verification

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package verification

type ClaimStore interface {
	ListByUser(ctx context.Context, userID string) ([]*Claim, error)
	ListByUserIDs(ctx context.Context, userIDs []string) ([]*Claim, error)
	ListByClaimName(ctx context.Context, userID string, claimName string) ([]*Claim, error)
	Get(ctx context.Context, userID string, claimName string, claimValue string) (*Claim, error)
	Create(ctx context.Context, claim *Claim) error
	Delete(ctx context.Context, id string) error
	DeleteAll(ctx context.Context, userID string) error
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

func (s *Service) GetIdentityVerificationStatus(ctx context.Context, i *identity.Info) ([]ClaimStatus, error) {
	claims, err := s.ClaimStore.ListByUser(ctx, i.UserID)
	if err != nil {
		return nil, err
	}

	verifiedClaims := make(map[claim]struct{})
	for _, c := range claims {
		verifiedClaims[claim{c.Name, c.Value}] = struct{}{}
	}

	return s.getVerificationStatus(i, verifiedClaims), nil
}

func (s *Service) GetVerificationStatuses(ctx context.Context, is []*identity.Info) (map[string][]ClaimStatus, error) {
	if len(is) == 0 {
		return nil, nil
	}

	idensByUserID := map[string][]*identity.Info{}
	for _, iden := range is {
		arr := idensByUserID[iden.UserID]
		idensByUserID[iden.UserID] = append(arr, iden)
	}

	userIDs := []string{}
	for userID := range idensByUserID {
		userIDs = append(userIDs, userID)
	}

	allClaims, err := s.ClaimStore.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	claimsByUserIDs := map[string][]*Claim{}
	for _, c := range allClaims {
		arr := claimsByUserIDs[c.UserID]
		claimsByUserIDs[c.UserID] = append(arr, c)
	}

	statuses := map[string][]ClaimStatus{}

	for userID, idens := range idensByUserID {
		claims, ok := claimsByUserIDs[userID]
		if !ok {
			claims = []*Claim{}
		}

		verifiedClaims := make(map[claim]struct{})
		for _, c := range claims {
			verifiedClaims[claim{c.Name, c.Value}] = struct{}{}
		}

		for _, i := range idens {
			if i.UserID != userID {
				panic("verification: expect all user ID is same")
			}
			statuses[i.ID] = s.getVerificationStatus(i, verifiedClaims)
		}
	}

	return statuses, nil
}

func (s *Service) GetAuthenticatorVerificationStatus(ctx context.Context, a *authenticator.Info) (AuthenticatorStatus, error) {
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

	_, err := s.ClaimStore.Get(ctx, a.UserID, claimName, claimValue)
	if errors.Is(err, ErrClaimUnverified) {
		return AuthenticatorStatusUnverified, nil
	} else if err != nil {
		return "", err
	}

	return AuthenticatorStatusVerified, nil
}

func (s *Service) GetClaims(ctx context.Context, userID string) ([]*Claim, error) {
	return s.ClaimStore.ListByUser(ctx, userID)
}

func (s *Service) GetClaimStatus(ctx context.Context, userID string, claimName model.ClaimName, claimValue string) (*ClaimStatus, error) {
	claims, err := s.ClaimStore.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	cfg := s.claimVerificationConfig(claimName)
	if cfg == nil {
		return nil, ErrUnsupportedClaim
	}

	verified := false
	var verifiedByChannel model.AuthenticatorOOBChannel
	for _, claim := range claims {
		if claim.Name == string(claimName) && claim.Value == claimValue {
			verified = true
			verifiedByChannel = claim.VerifiedByChannel()
			break
		}
	}

	return &ClaimStatus{
		Name:                       string(claimName),
		Value:                      claimValue,
		Verified:                   verified,
		RequiredToVerifyOnCreation: *cfg.Required,
		EndUserTriggerable:         *cfg.Enabled,
		VerifiedByChannel:          verifiedByChannel,
	}, nil
}

func (s *Service) AreUsersVerified(ctx context.Context, identitiesByUserIDs map[string][]*identity.Info) (map[string]bool, error) {
	allIdens := []*identity.Info{}
	for _, arr := range identitiesByUserIDs {
		allIdens = append(allIdens, arr...)
	}

	allStatuses, err := s.GetVerificationStatuses(ctx, allIdens)
	if err != nil {
		return nil, err
	}

	results := map[string]bool{}

	for userID, userIdens := range identitiesByUserIDs {
		statuses := []ClaimStatus{}
		for _, iden := range userIdens {
			if iden.UserID != userID {
				panic("verification: unexpected identity user ID")
			}
			statuses = append(statuses, allStatuses[iden.ID]...)
		}
		numVerifiable := 0
		numVerified := 0
		for _, claim := range statuses {
			if claim.IsVerifiable() {
				numVerifiable++
			}
			if claim.Verified {
				numVerified++
			}
		}

		switch s.Config.Criteria {
		case config.VerificationCriteriaAny:
			results[userID] = numVerifiable > 0 && numVerified >= 1
		case config.VerificationCriteriaAll:
			results[userID] = numVerifiable > 0 && numVerified == numVerifiable
		default:
			panic("verification: unknown criteria " + s.Config.Criteria)
		}
	}

	return results, nil
}

func (s *Service) IsUserVerified(ctx context.Context, identities []*identity.Info) (bool, error) {
	if len(identities) < 1 {
		return false, nil
	}
	userID := identities[0].UserID
	verifieds, err := s.AreUsersVerified(ctx, map[string][]*identity.Info{userID: identities})
	if err != nil {
		return false, err
	}
	if len(verifieds) != 1 {
		panic("verification: unexpected number of results returned")
	}
	return verifieds[userID], nil
}

func (s *Service) NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *Claim {
	return &Claim{
		ID:     uuid.New(),
		UserID: userID,
		Name:   claimName,
		Value:  claimValue,
	}
}

func (s *Service) MarkClaimVerified(ctx context.Context, claim *Claim) error {
	claims, err := s.GetClaims(ctx, claim.UserID)
	if err != nil {
		return err
	}
	for _, c := range claims {
		if c.Name == claim.Name && c.Value == claim.Value {
			return nil
		}
	}
	claim.CreatedAt = s.Clock.NowUTC()
	return s.ClaimStore.Create(ctx, claim)
}

func (s *Service) DeleteClaim(ctx context.Context, claim *Claim) error {
	return s.ClaimStore.Delete(ctx, claim.ID)
}

func (s *Service) ResetVerificationStatus(ctx context.Context, userID string) error {
	return s.ClaimStore.DeleteAll(ctx, userID)
}

func (s *Service) RemoveOrphanedClaims(ctx context.Context, userID string, identities []*identity.Info, authenticators []*authenticator.Info) error {
	claims, err := s.ClaimStore.ListByUser(ctx, userID)
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
		err = s.ClaimStore.Delete(ctx, claim.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
