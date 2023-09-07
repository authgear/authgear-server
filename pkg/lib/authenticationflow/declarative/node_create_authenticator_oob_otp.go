package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterNode(&NodeCreateAuthenticatorOOBOTP{})
}

type NodeCreateAuthenticatorOOBOTP struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeCreateAuthenticatorOOBOTP{}
var _ authflow.InputReactor = &NodeCreateAuthenticatorOOBOTP{}
var _ authflow.Milestone = &NodeCreateAuthenticatorOOBOTP{}
var _ MilestoneAuthenticationMethod = &NodeCreateAuthenticatorOOBOTP{}

func (*NodeCreateAuthenticatorOOBOTP) Kind() string {
	return "NodeCreateAuthenticatorOOBOTP"
}

func (*NodeCreateAuthenticatorOOBOTP) Milestone() {}
func (n *NodeCreateAuthenticatorOOBOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *NodeCreateAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)
	targetStepID := oneOf.TargetStep
	// If target step is specified, we do not need input to react.
	if targetStepID != "" {
		return nil, nil
	}

	return &InputTakeOOBOTPTarget{}, nil
}

func (n *NodeCreateAuthenticatorOOBOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)
	targetStepID := oneOf.TargetStep
	if targetStepID != "" {
		// Find the target step from the root.
		targetStepFlow, err := FindTargetStep(flows.Root, targetStepID)
		if err != nil {
			return nil, err
		}

		target, ok := targetStepFlow.Intent.(IntentSignupFlowStepAuthenticateTarget)
		if !ok {
			return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
				"target_step": targetStepID,
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
		return n.newNode(deps, oobOTPTarget)
	}

	var inputTakeOOBOTPTarget inputTakeOOBOTPTarget
	if authflow.AsInput(input, &inputTakeOOBOTPTarget) {
		oobOTPTarget := inputTakeOOBOTPTarget.GetTarget()
		return n.newNode(deps, oobOTPTarget)
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeCreateAuthenticatorOOBOTP) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}

func (n *NodeCreateAuthenticatorOOBOTP) newNode(deps *authflow.Dependencies, target string) (*authflow.Node, error) {
	spec := &authenticator.Spec{
		UserID: n.UserID,
		OOBOTP: &authenticator.OOBOTPSpec{},
	}

	switch n.Authentication {
	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		spec.Kind = model.AuthenticatorKindPrimary
		spec.Type = model.AuthenticatorTypeOOBEmail
		spec.OOBOTP.Email = target

	case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		spec.Kind = model.AuthenticatorKindPrimary
		spec.Type = model.AuthenticatorTypeOOBSMS
		spec.OOBOTP.Phone = target

	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		spec.Kind = model.AuthenticatorKindSecondary
		spec.Type = model.AuthenticatorTypeOOBEmail
		spec.OOBOTP.Email = target

	case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		spec.Kind = model.AuthenticatorKindSecondary
		spec.Type = model.AuthenticatorTypeOOBSMS
		spec.OOBOTP.Phone = target

	default:
		panic(fmt.Errorf("unexpected authentication method: %v", n.Authentication))
	}

	isDefault, err := authenticatorIsDefault(deps, n.UserID, spec.Kind)
	if err != nil {
		return nil, err
	}
	spec.IsDefault = isDefault

	authenticatorID := uuid.New()
	info, err := deps.Authenticators.NewWithAuthenticatorID(authenticatorID, spec)
	if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
		Authenticator: info,
	}), nil
}
