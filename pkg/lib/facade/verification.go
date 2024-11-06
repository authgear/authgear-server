package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
)

type AdminVerificationFacade struct {
	Verification VerificationService
	Coordinator  *Coordinator
}

func (v AdminVerificationFacade) GetClaims(ctx context.Context, userID string) ([]*verification.Claim, error) {
	return v.Verification.GetClaims(ctx, userID)
}

func (v AdminVerificationFacade) NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim {
	return v.Verification.NewVerifiedClaim(ctx, userID, claimName, claimValue)
}

func (v AdminVerificationFacade) MarkClaimVerified(ctx context.Context, claim *verification.Claim) error {
	return v.Coordinator.MarkClaimVerifiedByAdmin(ctx, claim)
}

func (v AdminVerificationFacade) DeleteClaim(ctx context.Context, claim *verification.Claim) error {
	return v.Coordinator.DeleteVerifiedClaimByAdmin(ctx, claim)
}

type WorkflowVerificationFacade struct {
	Verification VerificationService
}

func (v WorkflowVerificationFacade) GetClaimStatus(ctx context.Context, userID string, claimName model.ClaimName, claimValue string) (*verification.ClaimStatus, error) {
	return v.Verification.GetClaimStatus(ctx, userID, claimName, claimValue)
}

func (v WorkflowVerificationFacade) GetIdentityVerificationStatus(ctx context.Context, i *identity.Info) ([]verification.ClaimStatus, error) {
	return v.Verification.GetIdentityVerificationStatus(ctx, i)
}

func (v WorkflowVerificationFacade) NewVerifiedClaim(ctx context.Context, userID string, claimName string, claimValue string) *verification.Claim {
	return v.Verification.NewVerifiedClaim(ctx, userID, claimName, claimValue)
}

func (v WorkflowVerificationFacade) MarkClaimVerified(ctx context.Context, claim *verification.Claim) error {
	return v.Verification.MarkClaimVerified(ctx, claim)
}
