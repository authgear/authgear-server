package facade

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
)

type AdminVerificationFacade struct {
	Verification VerificationService
	Coordinator  *Coordinator
}

func (v AdminVerificationFacade) GetClaims(userID string) ([]*verification.Claim, error) {
	return v.Verification.GetClaims(userID)
}

func (v AdminVerificationFacade) NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim {
	return v.Verification.NewVerifiedClaim(userID, claimName, claimValue)
}

func (v AdminVerificationFacade) MarkClaimVerified(claim *verification.Claim) error {
	return v.Coordinator.MarkClaimVerifiedByAdmin(claim)
}

func (v AdminVerificationFacade) DeleteClaim(claim *verification.Claim) error {
	return v.Coordinator.DeleteVerifiedClaimByAdmin(claim)
}

type WorkflowVerificationFacade struct {
	Verification VerificationService
}

func (v WorkflowVerificationFacade) GetClaimStatus(userID string, claimName model.ClaimName, claimValue string) (*verification.ClaimStatus, error) {
	return v.Verification.GetClaimStatus(userID, claimName, claimValue)
}

func (v WorkflowVerificationFacade) GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error) {
	return v.Verification.GetIdentityVerificationStatus(i)
}

func (v WorkflowVerificationFacade) NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim {
	return v.Verification.NewVerifiedClaim(userID, claimName, claimValue)
}

func (v WorkflowVerificationFacade) MarkClaimVerified(claim *verification.Claim) error {
	return v.Verification.MarkClaimVerified(claim)
}
