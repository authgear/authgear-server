package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCheckIdentityConflict{})
}

type EdgeCheckIdentityConflict struct {
	NewIdentity *identity.Info
}

func (e *EdgeCheckIdentityConflict) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	dupeIdentity, err := ctx.Identities.CheckDuplicated(goCtx, e.NewIdentity)
	if err != nil && !identity.IsErrDuplicatedIdentity(err) {
		return nil, err
	}

	return &NodeCheckIdentityConflict{
		NewIdentity:        e.NewIdentity,
		DuplicatedIdentity: dupeIdentity,
	}, nil
}

type NodeCheckIdentityConflict struct {
	NewIdentity            *identity.Info                 `json:"new_identity"`
	DuplicatedIdentity     *identity.Info                 `json:"duplicated_identity"`
	IdentityConflictConfig *config.IdentityConflictConfig `json:"-"`
}

func (n *NodeCheckIdentityConflict) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	n.IdentityConflictConfig = ctx.Config.Identity.OnConflict
	return nil
}

func (n *NodeCheckIdentityConflict) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCheckIdentityConflict) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}

func (n *NodeCheckIdentityConflict) FillDetails(err error) error {
	spec := n.NewIdentity.ToSpec()
	otherSpec := n.DuplicatedIdentity.ToSpec()
	return identityFillDetails(err, &spec, &otherSpec)
}
