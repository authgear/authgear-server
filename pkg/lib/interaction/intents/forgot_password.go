package intents

import (
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

func (i *IntentForgotPassword) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeSelectIdentityBegin{}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentForgotPassword) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
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
