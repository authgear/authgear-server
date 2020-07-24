package intents

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentSignup{})
}

type IntentSignup struct {
}

func (i *IntentSignup) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	spec := nodes.EdgeDoCreateUser{}
	return spec.Instantiate(ctx, graph, i)
}

func (i *IntentSignup) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node.(type) {
	case *nodes.NodeDoCreateUser:
		return []newinteraction.Edge{
			&nodes.EdgeCreateIdentityBegin{},
		}, nil

	default:
		panic("interaction: unexpected node")
	}
}
