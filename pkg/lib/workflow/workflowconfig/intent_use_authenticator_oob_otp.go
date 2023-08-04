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
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentUseAuthenticatorOOBOTP{})
}

var IntentUseAuthenticatorOOBOTPSchema = validation.NewSimpleSchema(`{}`)

type IntentUseAuthenticatorOOBOTP struct {
	LoginFlow      string                              `json:"login_flow,omitempty"`
	JSONPointer    jsonpointer.T                       `json:"json_pointer,omitempty"`
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

var _ MilestoneAuthenticationMethod = &IntentUseAuthenticatorOOBOTP{}

func (*IntentUseAuthenticatorOOBOTP) Milestone() {}
func (n *IntentUseAuthenticatorOOBOTP) MilestoneAuthenticationMethod() (config.WorkflowAuthenticationMethod, bool) {
	return n.Authentication, true
}

var _ workflow.NodeSimple = &IntentUseAuthenticatorOOBOTP{}

func (*IntentUseAuthenticatorOOBOTP) Kind() string {
	return "workflowconfig.IntentUseAuthenticatorOOBOTP"
}

func (*IntentUseAuthenticatorOOBOTP) JSONSchema() *validation.SimpleSchema {
	return IntentUseAuthenticatorOOBOTPSchema
}

func (*IntentUseAuthenticatorOOBOTP) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{&InputTakeAuthenticatorID{}}, nil
	}

	_, authenticatorUsed := FindMilestone[MilestoneDoUseAuthenticator](workflows.Nearest)
	_, claimVerified := FindMilestone[MilestoneDoMarkClaimVerified](workflows.Nearest)

	switch {
	case authenticatorUsed && !claimVerified:
		// Verify claim
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

	if len(workflows.Nearest.Nodes) == 0 {
		var inputTakeAuthenticatorID inputTakeAuthenticatorID
		if workflow.AsInput(input, &inputTakeAuthenticatorID) {
			authenticatorID := inputTakeAuthenticatorID.GetAuthenticatorID()
			info, err := n.pickAuthenticator(deps, candidates, authenticatorID)
			if err != nil {
				return nil, err
			}

			return workflow.NewNodeSimple(&NodeDoUseAuthenticator{
				Authenticator: info,
			}), nil
		}
	}

	m, authenticatorUsed := FindMilestone[MilestoneDoUseAuthenticator](workflows.Nearest)
	_, claimVerified := FindMilestone[MilestoneDoMarkClaimVerified](workflows.Nearest)

	switch {
	case authenticatorUsed && !claimVerified:
		if nn, ok := m.MilestoneDoUseAuthenticator(); ok {
			info := nn.Authenticator
			claimName, claimValue := info.OOBOTP.ToClaimPair()
			return workflow.NewSubWorkflow(&IntentVerifyClaim{
				UserID:      n.UserID,
				Purpose:     otp.PurposeOOBOTP,
				MessageType: n.otpMessageType(info),
				ClaimName:   claimName,
				ClaimValue:  claimValue,
			}), nil
		}
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentUseAuthenticatorOOBOTP) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
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
