package loader

import (
	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type VerificationService interface {
	GetClaims(userID string) ([]*verification.Claim, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
	DeleteClaim(claimID string) error
}

type VerificationLoader struct {
	Verification VerificationService
}

func (l *VerificationLoader) Get(userID string) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		claims, err := l.Verification.GetClaims(userID)
		if err != nil {
			return nil, err
		}

		var models []model.Claim
		for _, c := range claims {
			models = append(models, model.Claim{
				Name:  c.Name,
				Value: c.Value,
			})
		}
		return models, nil
	})
}

func (l *VerificationLoader) SetVerified(userID string, claimName string, claimValue string, isVerified bool) *graphqlutil.Lazy {
	return graphqlutil.NewLazy(func() (interface{}, error) {
		claims, err := l.Verification.GetClaims(userID)
		if err != nil {
			return nil, err
		}

		// TODO(admin): use interaction for these operations
		if isVerified {
			for _, c := range claims {
				if c.Name == claimName && c.Value == claimValue {
					return nil, nil
				}
			}

			claim := l.Verification.NewVerifiedClaim(userID, claimName, claimValue)
			err = l.Verification.MarkClaimVerified(claim)
			if err != nil {
				return nil, err
			}

		} else {
			var claim *verification.Claim
			for _, c := range claims {
				if c.Name == claimName && c.Value == claimValue {
					claim = c
					break
				}
			}
			if claim == nil {
				return nil, nil
			}

			err = l.Verification.DeleteClaim(claim.ID)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
}
