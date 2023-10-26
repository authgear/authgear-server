package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentCreateAuthenticatorOOBOTP{})
}

type IntentCreateAuthenticatorOOBOTP struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentCreateAuthenticatorOOBOTP{}
var _ authflow.Milestone = &IntentCreateAuthenticatorOOBOTP{}
var _ MilestoneAuthenticationMethod = &IntentCreateAuthenticatorOOBOTP{}

func (*IntentCreateAuthenticatorOOBOTP) Kind() string {
	return "IntentCreateAuthenticatorOOBOTP"
}

func (*IntentCreateAuthenticatorOOBOTP) Milestone() {}
func (n *IntentCreateAuthenticatorOOBOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *IntentCreateAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	objectForOneOf, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(objectForOneOf)
	verificationRequired := oneOf.IsVerificationRequired()
	targetStepName := oneOf.TargetStep

	m, authenticatorSelected := authflow.FindMilestone[MilestoneDidSelectAuthenticator](flows.Nearest)
	claimVerifiedAlready := false
	_, claimVerifiedInThisFlow := authflow.FindMilestone[MilestoneDoMarkClaimVerified](flows.Nearest)
	_, created := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)

	if authenticatorSelected {
		info := m.MilestoneDidSelectAuthenticator()
		claimName, claimValue := info.OOBOTP.ToClaimPair()
		claimStatus, err := deps.Verification.GetClaimStatus(n.UserID, claimName, claimValue)
		if err != nil {
			return nil, err
		}
		claimVerifiedAlready = claimStatus.Verified
	}

	shouldVerifyInThisFlow := !claimVerifiedAlready && verificationRequired

	switch {
	case !authenticatorSelected:
		// If target step is specified, we do not need input to react.
		if targetStepName != "" {
			return nil, nil
		}

		return &InputSchemaTakeOOBOTPTarget{
			JSONPointer: n.JSONPointer,
		}, nil
	case shouldVerifyInThisFlow && !claimVerifiedInThisFlow:
		// Verify the claim
		return nil, nil
	case !created:
		// Create the authenticator
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (n *IntentCreateAuthenticatorOOBOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	objectForOneOf, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(objectForOneOf)
	verificationRequired := oneOf.IsVerificationRequired()
	targetStepName := oneOf.TargetStep

	m, authenticatorSelected := authflow.FindMilestone[MilestoneDidSelectAuthenticator](flows.Nearest)
	claimVerifiedAlready := false
	_, claimVerifiedInThisFlow := authflow.FindMilestone[MilestoneDoMarkClaimVerified](flows.Nearest)
	_, created := authflow.FindMilestone[MilestoneDoCreateAuthenticator](flows.Nearest)

	if authenticatorSelected {
		info := m.MilestoneDidSelectAuthenticator()
		claimName, claimValue := info.OOBOTP.ToClaimPair()
		claimStatus, err := deps.Verification.GetClaimStatus(n.UserID, claimName, claimValue)
		if err != nil {
			return nil, err
		}
		claimVerifiedAlready = claimStatus.Verified
	}

	shouldVerifyInThisFlow := !claimVerifiedAlready && verificationRequired

	switch {
	case !authenticatorSelected:
		if targetStepName != "" {
			// Find the target step from the root.
			targetStepFlow, err := authflow.FindTargetStep(flows.Root, targetStepName)
			if err != nil {
				return nil, err
			}

			target, ok := targetStepFlow.Intent.(IntentSignupFlowStepCreateAuthenticatorTarget)
			if !ok {
				return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
					"target_step": targetStepName,
				})
			}

			claims, err := target.GetOOBOTPClaims(ctx, deps, flows.Replace(targetStepFlow))
			if err != nil {
				return nil, err
			}

			var claimNames []model.ClaimName
			for claimName := range claims {
				claimNames = append(claimNames, claimName)
			}

			if len(claimNames) != 1 {
				// TODO(authflow): support create more than 1 OOB OTP authenticator?
				return nil, InvalidTargetStep.NewWithInfo("target_step does not contain exactly one claim for OOB-OTP", apierrors.Details{
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
				return nil, InvalidTargetStep.NewWithInfo("target_step contains unsupported claim for OOB-OTP", apierrors.Details{
					"claim_name": claimName,
				})
			}

			oobOTPTarget := claims[claimName]
			return n.newDidSelectAuthenticatorNode(deps, oobOTPTarget)
		}

		var inputTakeOOBOTPTarget inputTakeOOBOTPTarget
		if authflow.AsInput(input, &inputTakeOOBOTPTarget) {
			oobOTPTarget := inputTakeOOBOTPTarget.GetTarget()
			return n.newDidSelectAuthenticatorNode(deps, oobOTPTarget)
		}
	case shouldVerifyInThisFlow && !claimVerifiedInThisFlow:
		info := m.MilestoneDidSelectAuthenticator()
		claimName, claimValue := info.OOBOTP.ToClaimPair()
		purpose := otp.PurposeOOBOTP
		otpForm := getOTPForm(purpose, claimName, deps.Config.Authenticator.OOB.Email)
		return authflow.NewSubFlow(&IntentVerifyClaim{
			JSONPointer: n.JSONPointer,
			UserID:      n.UserID,
			Purpose:     purpose,
			MessageType: n.otpMessageType(),
			Form:        otpForm,
			ClaimName:   claimName,
			ClaimValue:  claimValue,
		}), nil
	case !created:
		info := m.MilestoneDidSelectAuthenticator()
		return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentCreateAuthenticatorOOBOTP) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}

func (i *IntentCreateAuthenticatorOOBOTP) otpMessageType() otp.MessageType {
	switch i.Authentication {
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		return otp.MessageTypeSetupPrimaryOOB
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		return otp.MessageTypeSetupPrimaryOOB
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		return otp.MessageTypeSetupSecondaryOOB
	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		return otp.MessageTypeSetupSecondaryOOB
	default:
		panic(fmt.Errorf("unexpected authentication method: %v", i.Authentication))
	}
}

func (n *IntentCreateAuthenticatorOOBOTP) newDidSelectAuthenticatorNode(deps *authflow.Dependencies, target string) (*authflow.Node, error) {
	info, err := createAuthenticator(deps, n.UserID, n.Authentication, target)
	if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(&NodeDidSelectAuthenticator{
		Authenticator: info,
	}), nil
}
