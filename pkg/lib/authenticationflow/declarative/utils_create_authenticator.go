package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func getCreateAuthenticatorOOBOTPTargetFromTargetStep(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	targetStepName string) (target string, isSkipped bool, err error) {
	targetStepFlow, err := authflow.FindTargetStep(flows.Root, targetStepName)
	if err != nil {
		return "", false, err
	}

	targetStep, ok := targetStepFlow.Intent.(IntentSignupFlowStepCreateAuthenticatorTarget)
	if !ok {
		return "", false, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
			"target_step": targetStepName,
		})
	}

	if targetStep.IsSkipped() {
		return "", true, nil
	}

	claims, err := targetStep.GetOOBOTPClaims(ctx, deps, flows.Replace(targetStepFlow))
	if err != nil {
		return "", false, err
	}

	var claimNames []model.ClaimName
	for claimName := range claims {
		claimNames = append(claimNames, claimName)
	}

	if len(claimNames) != 1 {
		// TODO(authflow): support create more than 1 OOB OTP authenticator?
		return "", false, InvalidTargetStep.NewWithInfo("target_step does not contain exactly one claim for OOB-OTP", apierrors.Details{
			"claims": claimNames,
		})
	}

	claimName := claimNames[0]
	switch claimName {
	case model.ClaimEmail:
		break
	case model.ClaimPhoneNumber:
		break
	default:
		return "", false, InvalidTargetStep.NewWithInfo("target_step contains unsupported claim for OOB-OTP", apierrors.Details{
			"claim_name": claimName,
		})
	}

	oobOTPTarget := claims[claimName]
	return oobOTPTarget, false, nil
}

func getCreateAuthenticatorOOBOTPTargetVerified(
	deps *authflow.Dependencies,
	userID string,
	claimName model.ClaimName,
	claimValue string) (bool, error) {
	claimStatus, err := deps.Verification.GetClaimStatus(userID, claimName, claimValue)
	if err != nil {
		return false, err
	}
	return claimStatus.Verified, nil
}
