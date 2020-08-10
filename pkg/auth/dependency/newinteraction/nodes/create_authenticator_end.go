package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorEnd{})
}

type EdgeCreateAuthenticatorEnd struct {
	Stage          newinteraction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeCreateAuthenticatorEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeCreateAuthenticatorEnd{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeCreateAuthenticatorEnd struct {
	Stage          newinteraction.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info              `json:"authenticators"`
}

func (n *NodeCreateAuthenticatorEnd) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorEnd) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
