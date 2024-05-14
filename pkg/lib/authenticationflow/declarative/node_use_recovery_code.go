package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeUseRecoveryCode{})
}

type NodeUseRecoveryCode struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
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

func (n *NodeUseRecoveryCode) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakeRecoveryCode{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
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
