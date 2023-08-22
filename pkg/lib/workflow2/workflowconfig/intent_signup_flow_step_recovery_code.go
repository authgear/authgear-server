package workflowconfig

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentSignupFlowStepRecoveryCode{})
}

type IntentSignupFlowStepRecoveryCodeData struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

func (*IntentSignupFlowStepRecoveryCodeData) Data() {}

type IntentSignupFlowStepRecoveryCode struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`

	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

func NewIntentSignupFlowStepRecoveryCode(deps *workflow.Dependencies, i *IntentSignupFlowStepRecoveryCode) *IntentSignupFlowStepRecoveryCode {
	i.RecoveryCodes = deps.MFA.GenerateRecoveryCodes()
	return i
}

var _ WorkflowStep = &IntentSignupFlowStepRecoveryCode{}

func (i *IntentSignupFlowStepRecoveryCode) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ workflow.Intent = &IntentSignupFlowStepRecoveryCode{}
var _ workflow.Boundary = &IntentSignupFlowStepRecoveryCode{}
var _ workflow.DataOutputer = &IntentSignupFlowStepRecoveryCode{}

func (*IntentSignupFlowStepRecoveryCode) Kind() string {
	return "workflowconfig.IntentSignupFlowStepRecoveryCode"
}

func (i *IntentSignupFlowStepRecoveryCode) Boundary() string {
	return i.JSONPointer.String()
}

func (*IntentSignupFlowStepRecoveryCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return &InputConfirmRecoveryCode{}, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowStepRecoveryCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		var inputConfirmRecoveryCode inputConfirmRecoveryCode
		if workflow.AsInput(input, &inputConfirmRecoveryCode) {
			return workflow.NewNodeSimple(&NodeDoReplaceRecoveryCode{
				UserID:        i.UserID,
				RecoveryCodes: i.RecoveryCodes,
			}), nil
		}
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentSignupFlowStepRecoveryCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.Data, error) {
	return &IntentSignupFlowStepRecoveryCodeData{
		RecoveryCodes: i.RecoveryCodes,
	}, nil
}
