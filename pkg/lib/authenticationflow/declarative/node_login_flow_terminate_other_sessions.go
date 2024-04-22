package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeLoginFlowTerminateOtherSessions{})
}

type NodeLoginFlowTerminateOtherSessions struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ authflow.NodeSimple = &NodeLoginFlowTerminateOtherSessions{}
var _ authflow.InputReactor = &NodeLoginFlowTerminateOtherSessions{}

func (*NodeLoginFlowTerminateOtherSessions) Kind() string {
	return "NodeLoginFlowTerminateOtherSessions"
}

func (n *NodeLoginFlowTerminateOtherSessions) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	return &InputConfirmTerminateOtherSessions{
		JSONPointer:    n.JSONPointer,
		FlowRootObject: flowRootObject,
	}, nil
}

func (n *NodeLoginFlowTerminateOtherSessions) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputConfirmTerminateOtherSessions inputConfirmTerminateOtherSessions
	if authflow.AsInput(input, &inputConfirmTerminateOtherSessions) {
		return authflow.NewNodeSimple(&NodeDidConfirmTerminateOtherSessions{}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
