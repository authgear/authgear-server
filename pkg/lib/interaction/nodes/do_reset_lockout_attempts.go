package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeDoResetLockoutAttempts{})
}

type EdgeDoResetLockoutAttempts struct {
}

func (e *EdgeDoResetLockoutAttempts) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	n := &NodeDoResetLockoutAttempts{}
	userID := ""
	authenticators := graph.GetUsedAuthenticators()
	types := []model.AuthenticatorType{}
	for _, a := range authenticators {
		userID = a.UserID
		types = append(types, a.Type)
	}
	if len(types) > 0 && userID != "" {
		ctx.Lockout.ClearAttempts(userID, types)
	}

	return n, nil
}

type NodeDoResetLockoutAttempts struct {
}

func (n *NodeDoResetLockoutAttempts) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoResetLockoutAttempts) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeDoResetLockoutAttempts) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
