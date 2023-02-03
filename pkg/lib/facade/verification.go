package facade

import "github.com/authgear/authgear-server/pkg/lib/feature/verification"

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
