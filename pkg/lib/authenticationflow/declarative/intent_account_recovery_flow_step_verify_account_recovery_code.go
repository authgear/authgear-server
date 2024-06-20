package declarative

import (
	"context"
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode{})
}

type IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData struct {
	TypedData
	MaskedDisplayName              string                 `json:"masked_display_name"`
	Channel                        AccountRecoveryChannel `json:"channel"`
	OTPForm                        AccountRecoveryOTPForm `json:"otp_form"`
	CodeLength                     int                    `json:"code_length,omitempty"`
	CanResendAt                    time.Time              `json:"can_resend_at,omitempty"`
	FailedAttemptRateLimitExceeded bool                   `json:"failed_attempt_rate_limit_exceeded"`
}

func NewIntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData(d IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData) IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData {
	d.Type = DataTypeAccountRecoveryVerifyCodeData
	return d
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
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaStepAccountRecoveryVerifyCode{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
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

		milestone, ok := i.findDestination(flows)
		if !ok {
			return nil, InvalidFlowConfig.New("NodeDoSendAccountRecoveryCode depends on MilestoneDoUseAccountRecoveryDestination")
		}
		dest := milestone.MilestoneDoUseAccountRecoveryDestination()

		nextNode := NewNodeDoSendAccountRecoveryCode(
			ctx,
			deps,
			i.FlowReference,
			i.JSONPointer,
			dest.TargetLoginID,
			dest.ForgotPasswordCodeKind(),
			dest.ForgotPasswordCodeChannel(),
		)
		// Ignore rate limit error on first entering the step.
		err := nextNode.Send(deps, true)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(nextNode), nil
	case 1:
		var inputStepAccountRecoveryVerifyCode inputStepAccountRecoveryVerifyCode
		if authflow.AsInput(input, &inputStepAccountRecoveryVerifyCode) {
			if inputStepAccountRecoveryVerifyCode.IsCode() {
				code := inputStepAccountRecoveryVerifyCode.GetCode()
				return i.verifyCode(deps, flows, code)
			}

			if inputStepAccountRecoveryVerifyCode.IsResend() {
				prevNode := flows.Nearest.Nodes[0].Simple
				switch prevNode.(type) {
				case *NodeDoSendAccountRecoveryCode:
					err := prevNode.(*NodeDoSendAccountRecoveryCode).Send(deps, false)
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

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) verifyCode(
	deps *authflow.Dependencies,
	flows authflow.Flows,
	code string,
) (*authflow.Node, error) {
	milestone, ok := i.findDestination(flows)
	if ok {
		dest := milestone.MilestoneDoUseAccountRecoveryDestination()
		_, err := deps.ResetPassword.VerifyCodeWithTarget(
			dest.TargetLoginID,
			code,
			dest.ForgotPasswordCodeChannel(),
			dest.ForgotPasswordCodeKind(),
		)
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

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) isRestored() bool {
	return isNodeRestored(i.JSONPointer, i.StartFrom)
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	milestone, ok := i.findDestination(flows)
	if ok {
		dest := milestone.MilestoneDoUseAccountRecoveryDestination()
		state, err := deps.ForgotPassword.InspectState(
			dest.TargetLoginID,
			dest.ForgotPasswordCodeChannel(),
			dest.ForgotPasswordCodeKind(),
		)
		if err != nil {
			return nil, err
		}
		codeLength := deps.ForgotPassword.CodeLength(
			dest.TargetLoginID,
			dest.ForgotPasswordCodeChannel(),
			dest.ForgotPasswordCodeKind(),
		)
		return NewIntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData(IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData{
			MaskedDisplayName:              dest.MaskedDisplayName,
			Channel:                        dest.Channel,
			OTPForm:                        dest.OTPForm,
			CodeLength:                     codeLength,
			CanResendAt:                    state.CanResendAt,
			FailedAttemptRateLimitExceeded: state.TooManyAttempts,
		}), nil
	} else {
		// MilestoneDoUseAccountRecoveryDestination might not exist, because the flow is restored
		return nil, nil
	}
}

func (i *IntentAccountRecoveryFlowStepVerifyAccountRecoveryCode) findDestination(flows authflow.Flows) (MilestoneDoUseAccountRecoveryDestination, bool) {
	ms := authflow.FindAllMilestones[MilestoneDoUseAccountRecoveryDestination](flows.Root)
	if len(ms) == 0 {
		return nil, false
	}
	// Otherwise use the first one we find.
	return ms[0], true
}
