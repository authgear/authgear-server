package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepResetPassword{})
}

type IntentAccountRecoveryFlowStepResetPassword struct {
	StepName    string        `json:"step_name,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.Intent = &IntentAccountRecoveryFlowStepResetPassword{}
var _ authflow.DataOutputer = &IntentAccountRecoveryFlowStepResetPassword{}

func (*IntentAccountRecoveryFlowStepResetPassword) Kind() string {
	return "IntentAccountRecoveryFlowStepResetPassword"
}

func (i *IntentAccountRecoveryFlowStepResetPassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaTakeNewPassword{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
		}, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentAccountRecoveryFlowStepResetPassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	milestone, ok := authflow.FindMilestone[MilestoneAccountRecoveryCode](flows.Root)
	if !ok {
		return nil, InvalidFlowConfig.New("IntentAccountRecoveryFlowStepResetPassword depends on MilestoneAccountRecoveryCode")
	}
	code := milestone.MilestoneAccountRecoveryCode()

	var inputTakeNewPassword inputTakeNewPassword
	if authflow.AsInput(input, &inputTakeNewPassword) {
		newPassword := inputTakeNewPassword.GetNewPassword()
		return authflow.NewNodeSimple(&NodeDoResetPassword{
			Code:        code,
			NewPassword: newPassword,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlowStepResetPassword) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewNewPasswordData(NewPasswordData{
		PasswordPolicy: NewPasswordPolicy(
			deps.FeatureConfig.Authenticator,
			deps.Config.Authenticator.Password.Policy,
		),
	}), nil
}
