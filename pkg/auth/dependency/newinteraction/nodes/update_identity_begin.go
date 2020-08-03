package nodes

import (
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
	IdentityID string `json:"identity_id"`
}

func (n *NodeUpdateIdentityBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUpdateIdentityBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	edges = append(edges, &EdgeUseIdentityLoginID{
		Mode:    UseIdentityLoginIDModeUpdate,
		Configs: ctx.Config.Identity.LoginID.Keys,
	})
	return edges, nil
}

func (n *NodeUpdateIdentityBegin) UpdateIdentityID() string {
	return n.IdentityID
}
