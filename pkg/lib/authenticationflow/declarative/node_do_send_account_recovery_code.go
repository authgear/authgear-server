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
	CodeChannel       forgotpassword.CodeChannel `json:"code_channel,omitempty"`
}

func NewNodeDoSendAccountRecoveryCode(
	ctx context.Context,
	deps *authflow.Dependencies,
	flowReference authflow.FlowReference,
	jsonPointer jsonpointer.T,
	targetLoginID string,
	codeKind forgotpassword.CodeKind,
	codeChannel forgotpassword.CodeChannel,
) *NodeDoSendAccountRecoveryCode {
	node := &NodeDoSendAccountRecoveryCode{
		ParentJSONPointer: jsonPointer,
		FlowReference:     flowReference,
		TargetLoginID:     targetLoginID,
		CodeKind:          codeKind,
		CodeChannel:       codeChannel,
	}

	return node
}

var _ authflow.NodeSimple = &NodeDoSendAccountRecoveryCode{}

func (*NodeDoSendAccountRecoveryCode) Kind() string {
	return "NodeDoSendAccountRecoveryCode"
}

func (n *NodeDoSendAccountRecoveryCode) Send(
	ctx context.Context,
	deps *authflow.Dependencies,
	ignoreRateLimitError bool,
) error {
	err := deps.ForgotPassword.SendCode(ctx, n.TargetLoginID, &forgotpassword.CodeOptions{
		AuthenticationFlowType:        string(n.FlowReference.Type),
		AuthenticationFlowName:        n.FlowReference.Name,
		AuthenticationFlowJSONPointer: n.ParentJSONPointer,
		Kind:                          n.CodeKind,
		Channel:                       n.CodeChannel,
	})

	if ignoreRateLimitError && deps.ForgotPassword.IsRateLimitError(err, n.TargetLoginID, n.CodeChannel, n.CodeKind) {
		// Ignore trigger cooldown rate limit error; continue the flow
	} else if errors.Is(err, forgotpassword.ErrUserNotFound) {
		// Do not tell user the user doen't exist
	} else if err != nil {
		return err
	}
	return nil
}
