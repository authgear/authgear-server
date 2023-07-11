package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	workflow.RegisterNode(&NodeCreateAuthenticatorOOBOTP{})
}

type NodeCreateAuthenticatorOOBOTP struct {
	SignupFlow     string                              `json:"signup_flow,omitempty"`
	JSONPointer    jsonpointer.T                       `json:"json_pointer,omitempty"`
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

func (*NodeCreateAuthenticatorOOBOTP) Kind() string {
	return "workflowconfig.NodeCreateAuthenticatorOOBOTP"
}

func (n *NodeCreateAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	current, err := signupFlowCurrent(deps, n.SignupFlow, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)
	targetStepID := oneOf.TargetStep
	// If target step is specified, we do not need input to react.
	if targetStepID != "" {
		return nil, nil
	}

	return []workflow.Input{&InputTakeOOBOTPTarget{}}, nil
}

func (n *NodeCreateAuthenticatorOOBOTP) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := signupFlowCurrent(deps, n.SignupFlow, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)
	targetStepID := oneOf.TargetStep
	if targetStepID != "" {
		// Find the target step from the root.
		targetStepWorkflow, err := FindTargetStep(workflows.Root, targetStepID)
		if err != nil {
			return nil, err
		}

		target, ok := targetStepWorkflow.Intent.(IntentSignupFlowStepAuthenticateTarget)
		if !ok {
			return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
				"target_step": targetStepID,
			})
		}

		claims, err := target.GetOOBOTPClaims(ctx, deps, workflows.Replace(targetStepWorkflow))
		if err != nil {
			return nil, err
		}

		var claimNames []model.ClaimName
		for claimName := range claims {
			claimNames = append(claimNames, claimName)
		}

		if len(claimNames) != 1 {
			// TODO(workflow): support create more than 1 OOB OTP authenticator?
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
	if workflow.AsInput(input, &inputTakeOOBOTPTarget) {
		oobOTPTarget := inputTakeOOBOTPTarget.GetTarget()
		return n.newNode(deps, oobOTPTarget)
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*NodeCreateAuthenticatorOOBOTP) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeCreateAuthenticatorOOBOTP) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorOOBOTP) oneOf(o config.WorkflowObject) *config.WorkflowSignupFlowOneOf {
	oneOf, ok := o.(*config.WorkflowSignupFlowOneOf)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return oneOf
}

func (n *NodeCreateAuthenticatorOOBOTP) newNode(deps *workflow.Dependencies, target string) (*workflow.Node, error) {
	spec := &authenticator.Spec{
		UserID: n.UserID,
		OOBOTP: &authenticator.OOBOTPSpec{},
	}

	switch n.Authentication {
	case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
		spec.Kind = model.AuthenticatorKindPrimary
		spec.Type = model.AuthenticatorTypeOOBEmail
		spec.OOBOTP.Email = target

	case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
		spec.Kind = model.AuthenticatorKindPrimary
		spec.Type = model.AuthenticatorTypeOOBSMS
		spec.OOBOTP.Phone = target

	case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
		spec.Kind = model.AuthenticatorKindSecondary
		spec.Type = model.AuthenticatorTypeOOBEmail
		spec.OOBOTP.Email = target

	case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
		spec.Kind = model.AuthenticatorKindSecondary
		spec.Type = model.AuthenticatorTypeOOBSMS
		spec.OOBOTP.Phone = target

	default:
		panic(fmt.Errorf("workflow: unexpected authentication method: %v", n.Authentication))
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

	return workflow.NewNodeSimple(&NodeDoCreateAuthenticator{
		Authenticator: info,
	}), nil
}
