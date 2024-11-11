package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeEnsurePasswordChange{})
}

type EdgeEnsurePasswordChange struct {
	Stage authn.AuthenticationStage
}

func (e *EdgeEnsurePasswordChange) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	authenticator, reason, ok := graph.GetRequireUpdateAuthenticator(e.Stage)
	if ok && authenticator.Type == model.AuthenticatorTypePassword {
		return &NodeChangePasswordBegin{
			Force:  true,
			Reason: reason,
			Stage:  e.Stage,
		}, nil
	}
	return &NodeEnsurePasswordChange{
		Stage: e.Stage,
	}, nil
}

type NodeEnsurePasswordChange struct {
	Stage authn.AuthenticationStage `json:"stage"`
}

func (n *NodeEnsurePasswordChange) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeEnsurePasswordChange) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeEnsurePasswordChange) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
