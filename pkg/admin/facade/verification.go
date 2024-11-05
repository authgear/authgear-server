package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
)

type VerificationService interface {
	NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim

	GetClaims(ctx context.Context, userID string) ([]*verification.Claim, error)
	MarkClaimVerified(ctx context.Context, claim *verification.Claim) error
	DeleteClaim(ctx context.Context, claim *verification.Claim) error
}

type VerificationFacade struct {
	Verification VerificationService
}

func (f *VerificationFacade) Get(ctx context.Context, userID string) ([]model.Claim, error) {
	claims, err := f.Verification.GetClaims(ctx, userID)
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
}

func (f *VerificationFacade) SetVerified(ctx context.Context, userID string, claimName string, claimValue string, isVerified bool) error {
	claims, err := f.Verification.GetClaims(ctx, userID)
	if err != nil {
		return err
	}

	// TODO(admin): use interaction for these operations
	if isVerified {
		for _, c := range claims {
			if c.Name == claimName && c.Value == claimValue {
				return nil
			}
		}

		claim := f.Verification.NewVerifiedClaim(ctx, userID, claimName, claimValue)
		err = f.Verification.MarkClaimVerified(ctx, claim)
		if err != nil {
			return err
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
			return nil
		}

		err = f.Verification.DeleteClaim(ctx, claim)
		if err != nil {
			return err
		}
	}

	return nil
}
