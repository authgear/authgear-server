package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{})
}

type IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData struct {
}

var _ authflow.Data = IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData{}

func (IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData) Data() {}

type IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepName    string        `json:"step_name,omitempty"`
}

var _ authflow.TargetStep = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) GetName() string {
	return i.StepName
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}
var _ authflow.DataOutputer = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}

func (*IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) Kind() string {
	return "IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode"
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		return &InputSchemaTakeAccountRecoveryCode{
			JSONPointer: i.JSONPointer,
		}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAccountRecoveryCode inputTakeAccountRecoveryCode
		if authflow.AsInput(input, &inputTakeAccountRecoveryCode) {
			code := inputTakeAccountRecoveryCode.GetAccountRecoveryCode()
			_, err := deps.ResetPassword.VerifyCode(code)
			if err != nil {
				return nil, err
			}
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData{}, nil
}
