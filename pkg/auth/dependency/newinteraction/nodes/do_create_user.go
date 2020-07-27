package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/uuid"
)

func init() {
	newinteraction.RegisterNode(&NodeDoCreateUser{})
}

type EdgeDoCreateUser struct {
}

func (e *EdgeDoCreateUser) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeDoCreateUser{
		NewUserID: uuid.New(),
	}, nil
}

type NodeDoCreateUser struct {
	NewUserID string `json:"new_user_id"`
}

func (n *NodeDoCreateUser) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		// User creation triggers hook, so run in a on commit effect
		// TODO(interaction): user metadata
		err := ctx.Users.Create(n.NewUserID, map[string]interface{}{}, graph.GetUserNewIdentities(), graph.GetUserNewAuthenticators())
		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoCreateUser) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeDoCreateUser) UserID() string {
	return n.NewUserID
}
