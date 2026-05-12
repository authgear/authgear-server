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

// accountRecoveryNoSendPrefix ("no-send:") is prepended to the username to form
// a TargetLoginID when username identification found the user but the user has
// no identity matching the selected channel. The resulting string is not a valid
// email address (no local@domain structure) and does not start with "+" so it
// cannot be an E.164 phone number, meaning SendCode always hits its
// generateDummyOTP path: no message is dispatched, but rate limits and
// cooldowns are still charged per username.
const accountRecoveryNoSendPrefix = "no-send:"

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
		AuthenticationFlowID:          authflow.GetFlowID(ctx),
		AuthenticationFlowType:        string(n.FlowReference.Type),
		AuthenticationFlowName:        n.FlowReference.Name,
		AuthenticationFlowJSONPointer: n.ParentJSONPointer,
		Kind:                          n.CodeKind,
		Channel:                       n.CodeChannel,
	})

	if ignoreRateLimitError && deps.ForgotPassword.IsRateLimitError(err, n.TargetLoginID, n.CodeChannel, n.CodeKind, authflow.GetFlowID(ctx)) {
		// Ignore trigger cooldown rate limit error; continue the flow
	} else if errors.Is(err, forgotpassword.ErrUserNotFound) {
		// Do not tell user the user doen't exist
	} else if err != nil {
		return err
	}
	return nil
}
