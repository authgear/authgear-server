package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeEnsurePasswordChange{})
}

type EdgeEnsurePasswordChange struct {
	Stage authn.AuthenticationStage
}

func (e *EdgeEnsurePasswordChange) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	authenticator, ok := graph.GetRequireUpdateAuthenticator(e.Stage)
	if ok && authenticator.Type == authn.AuthenticatorTypePassword {
		return &NodeChangePasswordBegin{
			Stage: e.Stage,
		}, nil
	}
	return &NodeEnsurePasswordChange{
		Stage: e.Stage,
	}, nil
}

type NodeEnsurePasswordChange struct {
	Stage authn.AuthenticationStage `json:"stage"`
}

func (n *NodeEnsurePasswordChange) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeEnsurePasswordChange) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeEnsurePasswordChange) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
