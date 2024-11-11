package intents

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentForgotPassword{})
}

type IntentForgotPassword struct{}

func NewIntentForgotPassword() *IntentForgotPassword {
	return &IntentForgotPassword{}
}

func (i *IntentForgotPassword) InstantiateRootNode(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeSelectIdentityBegin{}
	return edge.Instantiate(goCtx, ctx, graph, i)
}

func (i *IntentForgotPassword) DeriveEdgesForNode(goCtx context.Context, graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeSelectIdentityEnd:
		return []interaction.Edge{
			&nodes.EdgeForgotPasswordBegin{
				IdentityInfo: node.IdentityInfo,
			},
		}, nil
	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
