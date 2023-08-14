package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentUseAuthenticatorOOBOTP{})
}

type IntentUseAuthenticatorOOBOTP struct {
	LoginFlow      string                              `json:"login_flow,omitempty"`
	JSONPointer    jsonpointer.T                       `json:"json_pointer,omitempty"`
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

var _ workflow.Intent = &IntentUseAuthenticatorOOBOTP{}
var _ workflow.Milestone = &IntentUseAuthenticatorOOBOTP{}
var _ MilestoneAuthenticationMethod = &IntentUseAuthenticatorOOBOTP{}
var _ workflow.DataOutputer = &IntentUseAuthenticatorOOBOTP{}

func (*IntentUseAuthenticatorOOBOTP) Kind() string {
	return "workflowconfig.IntentUseAuthenticatorOOBOTP"
}

func (*IntentUseAuthenticatorOOBOTP) Milestone() {}
func (n *IntentUseAuthenticatorOOBOTP) MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return n.Authentication
}

func (*IntentUseAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	_, authenticatorSelected := workflow.FindMilestone[MilestoneDidSelectAuthenticator](workflows.Nearest)
	_, claimVerified := workflow.FindMilestone[MilestoneDoMarkClaimVerified](workflows.Nearest)
	_, authenticatorVerified := workflow.FindMilestone[MilestoneDidVerifyAuthenticator](workflows.Nearest)

	switch {
	case !authenticatorSelected:
		return []workflow.Input{&InputTakeAuthenticatorID{}}, nil
	case !claimVerified:
		// Verify the claim
		return nil, nil
	case !authenticatorVerified:
		// Achieve the milestone.
		return nil, nil
	default:
		return nil, workflow.ErrEOF
	}
}

func (n *IntentUseAuthenticatorOOBOTP) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	candidates, err := n.getCandidates(ctx, deps, workflows)
	if err != nil {
		return nil, err
	}

	m, authenticatorSelected := workflow.FindMilestone[MilestoneDidSelectAuthenticator](workflows.Nearest)
	_, claimVerified := workflow.FindMilestone[MilestoneDoMarkClaimVerified](workflows.Nearest)
	_, authenticatorVerified := workflow.FindMilestone[MilestoneDidVerifyAuthenticator](workflows.Nearest)

	switch {
	case !authenticatorSelected:
		var inputTakeAuthenticatorID inputTakeAuthenticatorID
		if workflow.AsInput(input, &inputTakeAuthenticatorID) {
			authenticatorID := inputTakeAuthenticatorID.GetAuthenticatorID()
			info, err := n.pickAuthenticator(deps, candidates, authenticatorID)
			if err != nil {
				return nil, err
			}

			return workflow.NewNodeSimple(&NodeDidSelectAuthenticator{
				Authenticator: info,
			}), nil
		}
	case !claimVerified:
		info := m.MilestoneDidSelectAuthenticator()
		claimName, claimValue := info.OOBOTP.ToClaimPair()
		return workflow.NewSubWorkflow(&IntentVerifyClaim{
			UserID:      n.UserID,
			Purpose:     otp.PurposeOOBOTP,
			MessageType: n.otpMessageType(info),
			ClaimName:   claimName,
			ClaimValue:  claimValue,
		}), nil
	case !authenticatorVerified:
		info := m.MilestoneDidSelectAuthenticator()
		return workflow.NewNodeSimple(&NodeDidVerifyAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (n *IntentUseAuthenticatorOOBOTP) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	candidates, err := n.getCandidates(ctx, deps, workflows)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"candidates": candidates,
	}, nil
}

func (*IntentUseAuthenticatorOOBOTP) oneOf(o config.WorkflowObject) *config.WorkflowLoginFlowOneOf {
	oneOf, ok := o.(*config.WorkflowLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return oneOf
}

func (n *IntentUseAuthenticatorOOBOTP) getCandidates(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]authenticator.Candidate, error) {

	current, err := loginFlowCurrent(deps, n.LoginFlow, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	var candidates []authenticator.Candidate

	oneOf := n.oneOf(current)
	targetStepID := oneOf.TargetStep
	if targetStepID != "" {
		// Find the target step from the root.
		targetStepWorkflow, err := FindTargetStep(workflows.Root, targetStepID)
		if err != nil {
			return nil, err
		}

		target, ok := targetStepWorkflow.Intent.(IntentLoginFlowStepAuthenticateTarget)
		if !ok {
			return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
				"target_step": targetStepID,
			})
		}

		identityInfo := target.GetIdentity(ctx, deps, workflows.Replace(targetStepWorkflow))

		candidates, err = getAuthenticationCandidatesOfIdentity(deps, identityInfo, n.Authentication)
		if err != nil {
			return nil, err
		}
	} else {
		candidates, err = getAuthenticationCandidatesOfUser(deps, n.UserID, []config.WorkflowAuthenticationMethod{n.Authentication})
		if err != nil {
			return nil, err
		}
	}

	return candidates, nil
}

func (n *IntentUseAuthenticatorOOBOTP) pickAuthenticator(deps *workflow.Dependencies, candidates []authenticator.Candidate, authenticatorID string) (*authenticator.Info, error) {
	for _, c := range candidates {
		id := c[authenticator.CandidateKeyAuthenticatorID].(string)
		if id == authenticatorID {
			info, err := deps.Authenticators.Get(authenticatorID)
			if err != nil {
				return nil, err
			}
			return info, nil
		}
	}

	return nil, InvalidAuthenticatorID.New("invalid authenticator ID")
}

func (*IntentUseAuthenticatorOOBOTP) otpMessageType(info *authenticator.Info) otp.MessageType {
	switch info.Kind {
	case model.AuthenticatorKindPrimary:
		return otp.MessageTypeAuthenticatePrimaryOOB
	case model.AuthenticatorKindSecondary:
		return otp.MessageTypeAuthenticateSecondaryOOB
	default:
		panic(fmt.Errorf("workflow: unexpected OOB OTP authenticator kind: %v", info.Kind))
	}
}
