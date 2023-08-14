package workflowconfig

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentSignupFlowStepRecoveryCode{})
}

type IntentSignupFlowStepRecoveryCode struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowStepRecoveryCode{}

func (i *IntentSignupFlowStepRecoveryCode) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (*IntentSignupFlowStepRecoveryCode) Kind() string {
	return "workflowconfig.IntentSignupFlowStepRecoveryCode"
}

func (*IntentSignupFlowStepRecoveryCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Generate a new set of recovery codes.
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowStepRecoveryCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		node := NewNodeGenerateRecoveryCode(deps, &NodeGenerateRecoveryCode{
			UserID: i.UserID,
		})
		return workflow.NewNodeSimple(node), nil
	}

	return nil, workflow.ErrIncompatibleInput
}
