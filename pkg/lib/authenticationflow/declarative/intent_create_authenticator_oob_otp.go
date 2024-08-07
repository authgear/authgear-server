package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentCreateAuthenticatorOOBOTP{})
}

type IntentCreateAuthenticatorOOBOTP struct {
	JSONPointer            jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID                 string                                  `json:"user_id,omitempty"`
	IsUpdatingExistingUser bool                                    `json:"is_updating_existing_user,omitempty"`
	Authentication         config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentCreateAuthenticatorOOBOTP{}
var _ authflow.Milestone = &IntentCreateAuthenticatorOOBOTP{}
var _ MilestoneAuthenticationMethod = &IntentCreateAuthenticatorOOBOTP{}
var _ MilestoneFlowCreateAuthenticator = &IntentCreateAuthenticatorOOBOTP{}

func (*IntentCreateAuthenticatorOOBOTP) Kind() string {
	return "IntentCreateAuthenticatorOOBOTP"
}

func (*IntentCreateAuthenticatorOOBOTP) Milestone() {}
func (*IntentCreateAuthenticatorOOBOTP) MilestoneFlowCreateAuthenticator(flows authflow.Flows) (MilestoneDoCreateAuthenticator, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
}
func (n *IntentCreateAuthenticatorOOBOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}
func (i *IntentCreateAuthenticatorOOBOTP) MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error {
	i.UserID = newUserID
	i.IsUpdatingExistingUser = true

	// Skip creation was handled by the parent IntentSignupFlowStepCreateAuthenticator
	// So don't need to do it here

	milestoneVerifyClaim, milestoneVeriyClaimFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneVerifyClaim](flows)
	if ok {
		return milestoneVerifyClaim.MilestoneVerifyClaimUpdateUserID(deps, milestoneVeriyClaimFlows, newUserID)
	}

	return nil
}

func (n *IntentCreateAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	objectForOneOf, err := authflow.FlowObject(flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	var verificationRequired bool
	var targetStepName string

	switch objectForOneOf.(type) {
	case *config.AuthenticationFlowSignupFlowOneOf:
		oneOf := objectForOneOf.(*config.AuthenticationFlowSignupFlowOneOf)
		verificationRequired = oneOf.IsVerificationRequired()
		targetStepName = oneOf.TargetStep
	case *config.AuthenticationFlowLoginFlowOneOf:
		oneOf := objectForOneOf.(*config.AuthenticationFlowLoginFlowOneOf)
		verificationRequired = true
		targetStepName = oneOf.TargetStep
	default:
		panic(fmt.Errorf("unexpected flow object: %T", objectForOneOf))
	}

	m, _, authenticatorSelected := authflow.FindMilestoneInCurrentFlow[MilestoneDidSelectAuthenticator](flows)
	claimVerifiedAlready := false
	_, _, claimVerifiedInThisFlow := authflow.FindMilestoneInCurrentFlow[MilestoneVerifyClaim](flows)
	_, _, created := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)

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

		isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
		if err != nil {
			return nil, err
		}

		return &InputSchemaTakeOOBOTPTarget{
			FlowRootObject:          flowRootObject,
			JSONPointer:             n.JSONPointer,
			IsBotProtectionRequired: isBotProtectionRequired,
			BotProtectionCfg:        deps.Config.BotProtection,
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
	rootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	objectForOneOf, err := authflow.FlowObject(rootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	var verificationRequired bool
	var targetStepName string

	switch objectForOneOf.(type) {
	case *config.AuthenticationFlowSignupFlowOneOf:
		oneOf := objectForOneOf.(*config.AuthenticationFlowSignupFlowOneOf)
		verificationRequired = oneOf.IsVerificationRequired()
		targetStepName = oneOf.TargetStep
	case *config.AuthenticationFlowLoginFlowOneOf:
		oneOf := objectForOneOf.(*config.AuthenticationFlowLoginFlowOneOf)
		verificationRequired = true
		targetStepName = oneOf.TargetStep
	default:
		panic(fmt.Errorf("unexpected flow object: %T", objectForOneOf))
	}

	m, _, authenticatorSelected := authflow.FindMilestoneInCurrentFlow[MilestoneDidSelectAuthenticator](flows)
	claimVerifiedAlready := false
	_, _, claimVerifiedInThisFlow := authflow.FindMilestoneInCurrentFlow[MilestoneVerifyClaim](flows)
	_, _, created := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)

	if authenticatorSelected {
		info := m.MilestoneDidSelectAuthenticator()
		claimName, claimValue := info.OOBOTP.ToClaimPair()
		verified, err := getCreateAuthenticatorOOBOTPTargetVerified(deps, n.UserID, claimName, claimValue)
		if err != nil {
			return nil, err
		}
		claimVerifiedAlready = verified
	}

	shouldVerifyInThisFlow := !claimVerifiedAlready && verificationRequired

	switch {
	case !authenticatorSelected:
		if targetStepName != "" {
			oobOTPTarget, _, err := getCreateAuthenticatorOOBOTPTargetFromTargetStep(ctx, deps, flows, targetStepName)
			if err != nil {
				return nil, err
			}
			if oobOTPTarget == "" {
				panic(fmt.Errorf("unexpected: oob otp target is empty"))
			}
			return n.newDidSelectAuthenticatorNode(deps, oobOTPTarget)
		}

		var inputTakeOOBOTPTarget inputTakeOOBOTPTarget
		if authflow.AsInput(input, &inputTakeOOBOTPTarget) {
			var bpSpecialErr error
			bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input)
			if err != nil {
				return nil, err
			}
			oobOTPTarget := inputTakeOOBOTPTarget.GetTarget()
			node, err := n.newDidSelectAuthenticatorNode(deps, oobOTPTarget)
			return node, errors.Join(bpSpecialErr, err)
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
