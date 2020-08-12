package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	newinteraction.RegisterNode(&NodeCheckIdentityConflict{})
}

type EdgeCheckIdentityConflict struct {
	NewIdentity *identity.Info
}

func (e *EdgeCheckIdentityConflict) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
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

func (n *NodeCheckIdentityConflict) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	n.IdentityConflictConfig = ctx.Config.Identity.OnConflict
	return nil
}

func (n *NodeCheckIdentityConflict) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCheckIdentityConflict) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
