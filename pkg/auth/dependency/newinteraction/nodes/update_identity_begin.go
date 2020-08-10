package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeUpdateIdentityBegin{})
}

type EdgeUpdateIdentityBegin struct {
	IdentityID string
}

func (e *EdgeUpdateIdentityBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeUpdateIdentityBegin{IdentityID: e.IdentityID}, nil
}

type NodeUpdateIdentityBegin struct {
	IdentityID  string                    `json:"identity_id"`
	LoginIDKeys []config.LoginIDKeyConfig `json:"-"`
}

func (n *NodeUpdateIdentityBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	n.LoginIDKeys = ctx.Config.Identity.LoginID.Keys
	return nil
}

func (n *NodeUpdateIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUpdateIdentityBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return n.deriveEdges(), nil
}

func (n *NodeUpdateIdentityBegin) deriveEdges() []newinteraction.Edge {
	var edges []newinteraction.Edge
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
