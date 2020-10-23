package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorEnd{})
}

type EdgeCreateAuthenticatorEnd struct {
	Stage          interaction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeCreateAuthenticatorEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeCreateAuthenticatorEnd{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeCreateAuthenticatorEnd struct {
	Stage          interaction.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info           `json:"authenticators"`
}

func (n *NodeCreateAuthenticatorEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
