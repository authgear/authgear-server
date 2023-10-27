package declarative

import (
	"context"
	"errors"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{})
}

type IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	StartFrom     jsonpointer.T          `json:"start_from,omitempty"`
}

var _ authflow.TargetStep = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}
var _ authflow.Instantiator = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) Instantiate(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
) error {
	if i.isRestored() {
		// We don't want to send the code again if this step was restored
		return nil
	}
	milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryDestination](flows.Root)
	if !ok {
		return InvalidFlowConfig.New("IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode depends on MilestoneDoUseAccountRecoveryDestination")
	}
	err := deps.ForgotPassword.SendCode(milestone.MilestoneDoUseAccountRecoveryDestination(), &forgotpassword.CodeOptions{
		AuthenticationFlowType:        string(i.FlowReference.Type),
		AuthenticationFlowName:        i.FlowReference.Name,
		AuthenticationFlowJSONPointer: i.JSONPointer,
	})
	if err != nil && !errors.Is(err, forgotpassword.ErrUserNotFound) {
		return err
	}
	return nil
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) GetName() string {
	return i.StepName
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}

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
			return authflow.NewNodeSimple(&NodeUseAccountRecoveryCode{Code: code}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) isRestored() bool {
	return isNodeRestored(i.JSONPointer, i.StartFrom)
}
