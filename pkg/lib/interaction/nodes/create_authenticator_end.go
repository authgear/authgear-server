package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorEnd{})
}

type EdgeCreateAuthenticatorEnd struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeCreateAuthenticatorEnd) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeCreateAuthenticatorEnd{
		Stage:          e.Stage,
		Authenticators: e.Authenticators,
	}, nil
}

type NodeCreateAuthenticatorEnd struct {
	Stage          authn.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info     `json:"authenticators"`
}

func (n *NodeCreateAuthenticatorEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
