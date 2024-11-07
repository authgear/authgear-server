package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeUpdateIdentityBegin{})
}

type EdgeUpdateIdentityBegin struct {
	IdentityID string
}

func (e *EdgeUpdateIdentityBegin) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeUpdateIdentityBegin{IdentityID: e.IdentityID}, nil
}

type NodeUpdateIdentityBegin struct {
	IdentityID  string                    `json:"identity_id"`
	LoginIDKeys []config.LoginIDKeyConfig `json:"-"`
}

func (n *NodeUpdateIdentityBegin) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	n.LoginIDKeys = ctx.Config.Identity.LoginID.Keys
	return nil
}

func (n *NodeUpdateIdentityBegin) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUpdateIdentityBegin) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.deriveEdges(), nil
}

func (n *NodeUpdateIdentityBegin) deriveEdges() []interaction.Edge {
	var edges []interaction.Edge
	edges = append(edges, &EdgeUseIdentityLoginID{
		Mode:    UseIdentityLoginIDModeUpdate,
		Configs: n.LoginIDKeys,
	})
	return edges
}

func (n *NodeUpdateIdentityBegin) UpdateIdentityID() string {
	return n.IdentityID
}

func (n *NodeUpdateIdentityBegin) GetIdentityCandidates() []identity.Candidate {
	var candidates []identity.Candidate
	for _, e := range n.deriveEdges() {
		if e, ok := e.(interface{ GetIdentityCandidates() []identity.Candidate }); ok {
			candidates = append(candidates, e.GetIdentityCandidates()...)
		}
	}
	return candidates
}
