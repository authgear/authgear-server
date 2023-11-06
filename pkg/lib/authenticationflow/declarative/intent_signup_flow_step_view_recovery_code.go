package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepViewRecoveryCode{})
}

type intentSignupFlowStepViewRecoveryCodeData struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

var _ authflow.Data = intentSignupFlowStepViewRecoveryCodeData{}

func (intentSignupFlowStepViewRecoveryCodeData) Data() {}

type IntentSignupFlowStepViewRecoveryCode struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepName    string        `json:"step_name,omitempty"`
	UserID      string        `json:"user_id,omitempty"`

	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

func NewIntentSignupFlowStepViewRecoveryCode(deps *authflow.Dependencies, i *IntentSignupFlowStepViewRecoveryCode) *IntentSignupFlowStepViewRecoveryCode {
	i.RecoveryCodes = deps.MFA.GenerateRecoveryCodes()
	return i
}

var _ authflow.Intent = &IntentSignupFlowStepViewRecoveryCode{}
var _ authflow.DataOutputer = &IntentSignupFlowStepViewRecoveryCode{}

func (*IntentSignupFlowStepViewRecoveryCode) Kind() string {
	return "IntentSignupFlowStepViewRecoveryCode"
}

func (i *IntentSignupFlowStepViewRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return &InputConfirmRecoveryCode{
			JSONPointer: i.JSONPointer,
		}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentSignupFlowStepViewRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
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

func (i *IntentSignupFlowStepViewRecoveryCode) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return intentSignupFlowStepViewRecoveryCodeData{
		RecoveryCodes: i.RecoveryCodes,
	}, nil
}
