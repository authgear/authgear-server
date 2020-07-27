package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/user"
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
	var u *user.User

	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		var err error
		u, err = ctx.Users.Create(n.NewUserID, map[string]interface{}{})
		return err
	}))
	if err != nil {
		return err
	}

	err = perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		return ctx.Users.AfterCreate(u, graph.GetUserNewIdentities(), graph.GetUserNewAuthenticators())
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
