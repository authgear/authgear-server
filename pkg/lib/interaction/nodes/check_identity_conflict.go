package nodes

import (
	"errors"

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

func (e *EdgeCheckIdentityConflict) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	dupeIdentity, err := ctx.Identities.CheckDuplicated(e.NewIdentity)
	if err != nil && !errors.Is(err, identity.ErrIdentityAlreadyExists) {
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

func (n *NodeCheckIdentityConflict) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	n.IdentityConflictConfig = ctx.Config.Identity.OnConflict
	return nil
}

func (n *NodeCheckIdentityConflict) Apply(perform func(eff interaction.Effect) error, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCheckIdentityConflict) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
