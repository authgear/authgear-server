package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeIdentifyWithIDToken{})
}

type NodeIdentifyWithIDToken struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.NodeSimple = &NodeIdentifyWithIDToken{}
var _ authflow.Milestone = &NodeIdentifyWithIDToken{}
var _ MilestoneIdentificationMethod = &NodeIdentifyWithIDToken{}
var _ authflow.InputReactor = &NodeIdentifyWithIDToken{}

func (*NodeIdentifyWithIDToken) Kind() string {
	return "NodeIdentifyWithIDToken"
}

func (*NodeIdentifyWithIDToken) Milestone() {}
func (n *NodeIdentifyWithIDToken) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *NodeIdentifyWithIDToken) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch {
	case authflow.GetIDToken(ctx) != "":
		// Special case: if id_token is available, use it automatically.
		return nil, nil
	default:
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaTakeIDToken{
			FlowRootObject: flowRootObject,
			JSONPointer:    n.JSONPointer,
		}, nil
	}
}

func (n *NodeIdentifyWithIDToken) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	proceed := func(idToken string) (*authflow.Node, error) {
		n, err := NewNodeDoUseIDToken(ctx, deps, flows, &NodeDoUseIDToken{
			IDToken: idToken,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), nil
	}

	switch {
	case authflow.GetIDToken(ctx) != "":
		idToken := authflow.GetIDToken(ctx)
		return proceed(idToken)
	default:
		var inputTakeIDToken inputTakeIDToken
		if authflow.AsInput(input, &inputTakeIDToken) {
			idToken := inputTakeIDToken.GetIDToken()
			return proceed(idToken)
		}

		return nil, authflow.ErrIncompatibleInput
	}
}
