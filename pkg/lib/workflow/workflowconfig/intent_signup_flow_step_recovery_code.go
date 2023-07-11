package workflowconfig

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentSignupFlowStepRecoveryCode{})
}

var IntentSignupFlowStepRecoveryCodeSchema = validation.NewSimpleSchema(`{}`)

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

func (*IntentSignupFlowStepRecoveryCode) JSONSchema() *validation.SimpleSchema {
	return IntentSignupFlowStepRecoveryCodeSchema
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

func (i *IntentSignupFlowStepRecoveryCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentSignupFlowStepRecoveryCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
