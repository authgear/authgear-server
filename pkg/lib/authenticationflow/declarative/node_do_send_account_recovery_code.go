package declarative

import (
	"context"
	"errors"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
)

func init() {
	authflow.RegisterNode(&NodeDoSendAccountRecoveryCode{})
}

type NodeDoSendAccountRecoveryCode struct {
	ParentJSONPointer jsonpointer.T              `json:"parent_json_pointer,omitempty"`
	FlowReference     authflow.FlowReference     `json:"flow_reference,omitempty"`
	TargetLoginID     string                     `json:"target_login_id,omitempty"`
	CodeKind          forgotpassword.CodeKind    `json:"code_kind,omitempty"`
	PhoneChannel      forgotpassword.CodeChannel `json:"phone_channel,omitempty"`
}

func NewNodeDoSendAccountRecoveryCode(
	ctx context.Context,
	deps *authflow.Dependencies,
	flowReference authflow.FlowReference,
	jsonPointer jsonpointer.T,
	targetLoginID string,
	codeKind forgotpassword.CodeKind,
	phoneChannel forgotpassword.CodeChannel,
) *NodeDoSendAccountRecoveryCode {
	node := &NodeDoSendAccountRecoveryCode{
		ParentJSONPointer: jsonPointer,
		FlowReference:     flowReference,
		TargetLoginID:     targetLoginID,
		CodeKind:          codeKind,
		PhoneChannel:      phoneChannel,
	}

	return node
}

var _ authflow.NodeSimple = &NodeDoSendAccountRecoveryCode{}

func (*NodeDoSendAccountRecoveryCode) Kind() string {
	return "NodeDoSendAccountRecoveryCode"
}

func (n *NodeDoSendAccountRecoveryCode) Send(
	deps *authflow.Dependencies,
) error {
	err := deps.ForgotPassword.SendCode(n.TargetLoginID, &forgotpassword.CodeOptions{
		AuthenticationFlowType:        string(n.FlowReference.Type),
		AuthenticationFlowName:        n.FlowReference.Name,
		AuthenticationFlowJSONPointer: n.ParentJSONPointer,
		Kind:                          n.CodeKind,
		Channel:                       n.PhoneChannel,
	})
	if err != nil && !errors.Is(err, forgotpassword.ErrUserNotFound) {
		return err
	}
	return nil
}
