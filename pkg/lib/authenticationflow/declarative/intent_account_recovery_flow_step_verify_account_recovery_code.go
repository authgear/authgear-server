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
	MaskedDisplayName string                 `json:"masked_display_name"`
	Channel           AccountRecoveryChannel `json:"channel"`
	OTPForm           AccountRecoveryOTPForm `json:"otp_form"`
}

var _ authflow.Data = IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData{}

func (IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData) Data() {}

type IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	StartFrom     jsonpointer.T          `json:"start_from,omitempty"`
}

var _ authflow.TargetStep = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}
var _ authflow.DataOutputer = &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{}

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
	switch len(flows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return &InputSchemaStepAccountRecoveryVerifyCode{
			JSONPointer: i.JSONPointer,
		}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		if i.isRestored() {
			// We don't want to send the code again if this step was restored
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}
		nextNode, err := NewNodeDoSendAccountRecoveryCode(ctx, deps, flows, i.FlowReference, i.JSONPointer)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(nextNode), nil
	case 1:
		var inputStepAccountRecoveryVerifyCode inputStepAccountRecoveryVerifyCode
		if authflow.AsInput(input, &inputStepAccountRecoveryVerifyCode) {
			if inputStepAccountRecoveryVerifyCode.IsCode() {
				code := inputStepAccountRecoveryVerifyCode.GetCode()
				milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryDestination](flows.Root)
				if ok {
					dest := milestone.MilestoneDoUseAccountRecoveryDestination()
					_, err := deps.ResetPassword.VerifyCodeWithTarget(dest.TargetLoginID, code)
					if err != nil {
						return nil, err
					}
				} else {
					// MilestoneDoUseAccountRecoveryDestination might not exist, because the flow is restored
					_, err := deps.ResetPassword.VerifyCode(code)
					if err != nil {
						return nil, err
					}
				}
				return authflow.NewNodeSimple(&NodeUseAccountRecoveryCode{Code: code}), nil
			}

			if inputStepAccountRecoveryVerifyCode.IsResend() {
				prevNode := flows.Nearest.Nodes[0].Simple
				switch prevNode.(type) {
				case *NodeDoSendAccountRecoveryCode:
					err := prevNode.(*NodeDoSendAccountRecoveryCode).send(deps)
					if err != nil {
						return nil, err
					}
					return authflow.NewNodeSimple(prevNode), authflow.ErrUpdateNode
				}
			}

			return nil, authflow.ErrIncompatibleInput
		}

	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) isRestored() bool {
	return isNodeRestored(i.JSONPointer, i.StartFrom)
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryDestination](flows.Root)
	if ok {
		dest := milestone.MilestoneDoUseAccountRecoveryDestination()
		return &IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData{
			MaskedDisplayName: dest.MaskedDisplayName,
			Channel:           dest.Channel,
			OTPForm:           dest.OTPForm,
		}, nil
	} else {
		// MilestoneDoUseAccountRecoveryDestination might not exist, because the flow is restored
		return nil, nil
	}
}
