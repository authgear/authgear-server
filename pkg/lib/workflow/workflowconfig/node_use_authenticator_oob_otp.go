package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeUseAuthenticatorOOBOTP{})
}

type NodeUseAuthenticatorOOBOTP struct {
	LoginFlow      string                              `json:"login_flow,omitempty"`
	JSONPointer    jsonpointer.T                       `json:"json_pointer,omitempty"`
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

var _ MilestoneAuthenticationMethod = &NodeUseAuthenticatorOOBOTP{}

func (*NodeUseAuthenticatorOOBOTP) Milestone() {}
func (n *NodeUseAuthenticatorOOBOTP) MilestoneAuthenticationMethod() (config.WorkflowAuthenticationMethod, bool) {
	return n.Authentication, true
}

var _ workflow.NodeSimple = &NodeUseAuthenticatorOOBOTP{}

func (*NodeUseAuthenticatorOOBOTP) Kind() string {
	return "workflowconfig.NodeUseAuthenticatorOOBOTP"
}

func (*NodeUseAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeAuthenticatorID{}}, nil
}

func (n *NodeUseAuthenticatorOOBOTP) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	candidates, err := n.getCandidates(ctx, deps, workflows)
	if err != nil {
		return nil, err
	}

	var inputTakeAuthenticatorID inputTakeAuthenticatorID
	if workflow.AsInput(input, &inputTakeAuthenticatorID) {
		authenticatorID := inputTakeAuthenticatorID.GetAuthenticatorID()
		_, err := n.pickAuthenticator(deps, candidates, authenticatorID)
		if err != nil {
			return nil, err
		}

		// FIXME(workflow): verify OTP.
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*NodeUseAuthenticatorOOBOTP) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (n *NodeUseAuthenticatorOOBOTP) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	candidates, err := n.getCandidates(ctx, deps, workflows)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"candidates": candidates,
	}, nil
}

func (*NodeUseAuthenticatorOOBOTP) oneOf(o config.WorkflowObject) *config.WorkflowLoginFlowOneOf {
	oneOf, ok := o.(*config.WorkflowLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return oneOf
}

func (n *NodeUseAuthenticatorOOBOTP) getCandidates(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]authenticator.Candidate, error) {

	current, err := loginFlowCurrent(deps, n.LoginFlow, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	candidates := []authenticator.Candidate{}

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

func (n *NodeUseAuthenticatorOOBOTP) pickAuthenticator(deps *workflow.Dependencies, candidates []authenticator.Candidate, authenticatorID string) (*authenticator.Info, error) {
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
