package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityEnd{})
}

type EdgeCreateIdentityEnd struct {
	IdentitySpec *identity.Spec
}

func (e *EdgeCreateIdentityEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	info, err := ctx.Identities.New(graph.MustGetUserID(), e.IdentitySpec)
	if err != nil {
		return nil, err
	}

	return &NodeCreateIdentityEnd{
		IdentitySpec: e.IdentitySpec,
		IdentityInfo: info,
	}, nil
}

type NodeCreateIdentityEnd struct {
	IdentitySpec *identity.Spec `json:"identity_spec"`
	IdentityInfo *identity.Info `json:"identity_info"`
}

func (n *NodeCreateIdentityEnd) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityEnd) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
