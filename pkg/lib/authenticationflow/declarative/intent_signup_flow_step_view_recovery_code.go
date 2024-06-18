package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepViewRecoveryCode{})
}

type IntentSignupFlowStepViewRecoveryCodeData struct {
	TypedData
	RecoveryCodes []string `json:"recovery_codes"`
}

func NewIntentSignupFlowStepViewRecoveryCodeData(d IntentSignupFlowStepViewRecoveryCodeData) IntentSignupFlowStepViewRecoveryCodeData {
	d.Type = DataTypeViewRecoveryCodeData
	return d
}

var _ authflow.Data = IntentSignupFlowStepViewRecoveryCodeData{}

func (IntentSignupFlowStepViewRecoveryCodeData) Data() {}

type IntentSignupFlowStepViewRecoveryCode struct {
	JSONPointer            jsonpointer.T `json:"json_pointer,omitempty"`
	StepName               string        `json:"step_name,omitempty"`
	UserID                 string        `json:"user_id,omitempty"`
	IsUpdatingExistingUser bool          `json:"is_updating_existing_user,omitempty"`

	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

func NewIntentSignupFlowStepViewRecoveryCode(deps *authflow.Dependencies, i *IntentSignupFlowStepViewRecoveryCode) *IntentSignupFlowStepViewRecoveryCode {
	i.RecoveryCodes = deps.MFA.GenerateRecoveryCodes()
	return i
}

var _ authflow.Intent = &IntentSignupFlowStepViewRecoveryCode{}
var _ authflow.DataOutputer = &IntentSignupFlowStepViewRecoveryCode{}
var _ authflow.Milestone = &IntentSignupFlowStepViewRecoveryCode{}
var _ MilestoneSwitchToExistingUser = &IntentSignupFlowStepViewRecoveryCode{}

func (*IntentSignupFlowStepViewRecoveryCode) Milestone() {}
func (i *IntentSignupFlowStepViewRecoveryCode) MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error {
	i.UserID = newUserID
	i.IsUpdatingExistingUser = true

	milestone, ok := authflow.FindFirstMilestone[MilestoneDoReplaceRecoveryCode](flows.Nearest)
	if ok {
		milestone.MilestoneDoReplaceRecoveryCodeUpdateUserID(newUserID)
	}

	return nil
}

func (*IntentSignupFlowStepViewRecoveryCode) Kind() string {
	return "IntentSignupFlowStepViewRecoveryCode"
}

func (i *IntentSignupFlowStepViewRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if !i.IsUpdatingExistingUser && len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputConfirmRecoveryCode{
			JSONPointer:    i.JSONPointer,
			FlowRootObject: flowRootObject,
		}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentSignupFlowStepViewRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	if !i.IsUpdatingExistingUser && len(flows.Nearest.Nodes) == 0 {
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
	return NewIntentSignupFlowStepViewRecoveryCodeData(IntentSignupFlowStepViewRecoveryCodeData{
		RecoveryCodes: i.RecoveryCodes,
	}), nil
}
