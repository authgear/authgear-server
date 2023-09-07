package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepRecoveryCode{})
}

type IntentSignupFlowStepRecoveryCodeData struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

var _ authflow.Data = &IntentSignupFlowStepRecoveryCodeData{}

func (*IntentSignupFlowStepRecoveryCodeData) Data() {}

type IntentSignupFlowStepRecoveryCode struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`

	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

func NewIntentSignupFlowStepRecoveryCode(deps *authflow.Dependencies, i *IntentSignupFlowStepRecoveryCode) *IntentSignupFlowStepRecoveryCode {
	i.RecoveryCodes = deps.MFA.GenerateRecoveryCodes()
	return i
}

var _ FlowStep = &IntentSignupFlowStepRecoveryCode{}

func (i *IntentSignupFlowStepRecoveryCode) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentSignupFlowStepRecoveryCode{}
var _ authflow.DataOutputer = &IntentSignupFlowStepRecoveryCode{}

func (*IntentSignupFlowStepRecoveryCode) Kind() string {
	return "IntentSignupFlowStepRecoveryCode"
}

func (*IntentSignupFlowStepRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return &InputConfirmRecoveryCode{}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentSignupFlowStepRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputConfirmRecoveryCode inputConfirmRecoveryCode
		if authflow.AsInput(input, &inputConfirmRecoveryCode) {
			return authflow.NewNodeSimple(&NodeDoReplaceRecoveryCode{
				UserID:        i.UserID,
				RecoveryCodes: i.RecoveryCodes,
			}), nil
		}
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentSignupFlowStepRecoveryCode) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return &IntentSignupFlowStepRecoveryCodeData{
		RecoveryCodes: i.RecoveryCodes,
	}, nil
}
