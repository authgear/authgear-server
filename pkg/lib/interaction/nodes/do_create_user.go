package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateUser{})
}

type EdgeDoCreateUser struct {
}

func (e *EdgeDoCreateUser) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeDoCreateUser{
		CreateUserID: uuid.New(),
	}, nil
}

type NodeDoCreateUser struct {
	CreateUserID string `json:"create_user_id"`
}

func (n *NodeDoCreateUser) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoCreateUser) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	var u *user.User

	err := perform(interaction.EffectRun(func(ctx *interaction.Context) error {
		var err error
		u, err = ctx.Users.Create(n.CreateUserID)
		return err
	}))
	if err != nil {
		return err
	}

	err = perform(interaction.EffectOnCommit(func(ctx *interaction.Context) error {
		// TODO(interaction): add verified claim eagerly?
		return ctx.Users.AfterCreate(u, graph.GetUserNewIdentities())
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoCreateUser) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}

func (n *NodeDoCreateUser) UserID() string {
	return n.CreateUserID
}

func (n *NodeDoCreateUser) NewUserID() string {
	return n.CreateUserID
}
