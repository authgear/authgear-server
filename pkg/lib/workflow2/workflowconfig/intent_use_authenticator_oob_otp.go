package workflowconfig

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentUseAuthenticatorOOBOTP{})
}

type IntentUseAuthenticatorOOBOTPData struct {
	Candidates []AuthenticationCandidate `json:"candidates"`
}

var _ workflow.Data = IntentUseAuthenticatorOOBOTPData{}

func (m IntentUseAuthenticatorOOBOTPData) Data() {}

type IntentUseAuthenticatorOOBOTP struct {
	LoginFlow         string                              `json:"login_flow,omitempty"`
	JSONPointer       jsonpointer.T                       `json:"json_pointer,omitempty"`
	JSONPointerToStep jsonpointer.T                       `json:"json_pointer_to_step,omitempty"`
	UserID            string                              `json:"user_id,omitempty"`
	Authentication    config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
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
		return []workflow.Input{&InputTakeAuthenticationCandidateIndex{}}, nil
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
	m, authenticatorSelected := workflow.FindMilestone[MilestoneDidSelectAuthenticator](workflows.Nearest)
	_, claimVerified := workflow.FindMilestone[MilestoneDoMarkClaimVerified](workflows.Nearest)
	_, authenticatorVerified := workflow.FindMilestone[MilestoneDidVerifyAuthenticator](workflows.Nearest)

	switch {
	case !authenticatorSelected:
		var inputTakeAuthenticationCandidateIndex inputTakeAuthenticationCandidateIndex
		if workflow.AsInput(input, &inputTakeAuthenticationCandidateIndex) {
			current, err := loginFlowCurrent(deps, n.LoginFlow, n.JSONPointerToStep)
			if err != nil {
				return nil, err
			}
			step := n.step(current)

			candidates, err := getAuthenticationCandidatesForStep(ctx, deps, workflows, n.UserID, step)
			if err != nil {
				return nil, err
			}

			index := inputTakeAuthenticationCandidateIndex.GetIndex()
			info, err := n.pickAuthenticator(deps, candidates, index)
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

func (n *IntentUseAuthenticatorOOBOTP) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.Data, error) {
	current, err := loginFlowCurrent(deps, n.LoginFlow, n.JSONPointerToStep)
	if err != nil {
		return nil, err
	}
	step := n.step(current)

	candidates, err := getAuthenticationCandidatesForStep(ctx, deps, workflows, n.UserID, step)
	if err != nil {
		return nil, err
	}

	return IntentUseAuthenticatorOOBOTPData{
		Candidates: candidates,
	}, nil
}

func (*IntentUseAuthenticatorOOBOTP) step(o config.WorkflowObject) *config.WorkflowLoginFlowStep {
	step, ok := o.(*config.WorkflowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (n *IntentUseAuthenticatorOOBOTP) pickAuthenticator(deps *workflow.Dependencies, candidates []AuthenticationCandidate, index int) (*authenticator.Info, error) {
	for idx, c := range candidates {
		if idx == index {
			id := c.AuthenticatorID
			info, err := deps.Authenticators.Get(id)
			if errors.Is(err, authenticator.ErrAuthenticatorNotFound) {
				// Break the loop and use the return statement at the end.
				break
			}
			if err != nil {
				return nil, err
			}
			return info, nil
		}
	}

	return nil, InvalidAuthenticationCandidateIndex.New("invalid authentication candidate index")
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
