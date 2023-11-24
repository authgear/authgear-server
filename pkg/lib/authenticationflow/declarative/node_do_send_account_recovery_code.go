package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
)

func init() {
	authflow.RegisterNode(&NodeDoSendAccountRecoveryCode{})
}

type NodeDoSendAccountRecoveryCode struct {
}

func NewNodeDoSendAccountRecoveryCode(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	flowReference authflow.FlowReference,
	jSONPointer jsonpointer.T,
	startFrom jsonpointer.T,
) (*NodeDoSendAccountRecoveryCode, error) {
	milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryDestination](flows.Root)
	if !ok {
		return nil, InvalidFlowConfig.New("NodeDoSendAccountRecoveryCode depends on MilestoneDoUseAccountRecoveryDestination")
	}
	destination := milestone.MilestoneDoUseAccountRecoveryDestination()
	err := deps.ForgotPassword.SendCode(destination.TargetLoginID, &forgotpassword.CodeOptions{
		AuthenticationFlowType:        string(flowReference.Type),
		AuthenticationFlowName:        flowReference.Name,
		AuthenticationFlowJSONPointer: jSONPointer,
		Kind:                          accountRecoveryOTPFormToForgotPasswordCodeKind(destination.OTPForm),
	})
	if err != nil && !errors.Is(err, forgotpassword.ErrUserNotFound) {
		return nil, err
	}
	return &NodeDoSendAccountRecoveryCode{}, nil
}

var _ authflow.NodeSimple = &NodeDoSendAccountRecoveryCode{}

func (*NodeDoSendAccountRecoveryCode) Kind() string {
	return "NodeDoSendAccountRecoveryCode"
}

func accountRecoveryOTPFormToForgotPasswordCodeKind(otpForm AccountRecoveryOTPForm) forgotpassword.CodeKind {
	switch otpForm {
	case AccountRecoveryOTPFormCode:
		return forgotpassword.CodeKindShortCode
	case AccountRecoveryOTPFormLink:
		return forgotpassword.CodeKindLink
	}
	panic(fmt.Sprintf("account recovery: unknown otp form %s", otpForm))
}
