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
	Destination       AccountRecoveryDestinationOptionInternal `json:"destination,omitempty"`
	ParentJSONPointer jsonpointer.T                            `json:"parent_json_pointer,omitempty"`
	FlowReference     authflow.FlowReference                   `json:"flow_reference,omitempty"`
}

func NewNodeDoSendAccountRecoveryCode(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	flowReference authflow.FlowReference,
	jsonPointer jsonpointer.T,
) (*NodeDoSendAccountRecoveryCode, error) {
	milestone, ok := authflow.FindMilestone[MilestoneDoUseAccountRecoveryDestination](flows.Root)
	if !ok {
		return nil, InvalidFlowConfig.New("NodeDoSendAccountRecoveryCode depends on MilestoneDoUseAccountRecoveryDestination")
	}
	destination := milestone.MilestoneDoUseAccountRecoveryDestination()

	node := &NodeDoSendAccountRecoveryCode{
		Destination:       *destination,
		ParentJSONPointer: jsonPointer,
		FlowReference:     flowReference,
	}

	err := node.send(deps)
	if err != nil {
		return nil, err
	}

	return node, nil
}

var _ authflow.NodeSimple = &NodeDoSendAccountRecoveryCode{}

func (*NodeDoSendAccountRecoveryCode) Kind() string {
	return "NodeDoSendAccountRecoveryCode"
}

func (n *NodeDoSendAccountRecoveryCode) send(
	deps *authflow.Dependencies,
) error {
	err := deps.ForgotPassword.SendCode(n.Destination.TargetLoginID, &forgotpassword.CodeOptions{
		AuthenticationFlowType:        string(n.FlowReference.Type),
		AuthenticationFlowName:        n.FlowReference.Name,
		AuthenticationFlowJSONPointer: n.ParentJSONPointer,
		Kind:                          accountRecoveryOTPFormToForgotPasswordCodeKind(n.Destination.OTPForm),
	})
	if err != nil && !errors.Is(err, forgotpassword.ErrUserNotFound) {
		return err
	}
	return nil
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
