package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeUseRecoveryCode{})
}

type NodeUseRecoveryCode struct {
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseRecoveryCode{}
var _ authflow.Milestone = &NodeUseRecoveryCode{}
var _ MilestoneAuthenticationMethod = &NodeUseRecoveryCode{}
var _ authflow.InputReactor = &NodeUseRecoveryCode{}

func (*NodeUseRecoveryCode) Kind() string {
	return "NodeUseRecoveryCode"
}

func (*NodeUseRecoveryCode) Milestone() {}
func (n *NodeUseRecoveryCode) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*NodeUseRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputTakeRecoveryCode{}, nil
}

func (n *NodeUseRecoveryCode) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeRecoveryCode inputTakeRecoveryCode
	if authflow.AsInput(input, &inputTakeRecoveryCode) {
		recoveryCode := inputTakeRecoveryCode.GetRecoveryCode()

		rc, err := deps.MFA.VerifyRecoveryCode(n.UserID, recoveryCode)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoConsumeRecoveryCode{
			RecoveryCode: rc,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
