package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeSelectIdentityEnd{})
}

type EdgeSelectIdentityEnd struct {
	IdentitySpec *identity.Spec
}

func (e *EdgeSelectIdentityEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	var info *identity.Info
	info, err := ctx.Identities.GetBySpec(e.IdentitySpec)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		// nolint: ineffassign
		err = nil
	} else if err != nil {
		return nil, err
	}

	return &NodeSelectIdentityEnd{
		IdentitySpec: e.IdentitySpec,
		IdentityInfo: info,
	}, nil
}

type NodeSelectIdentityEnd struct {
	IdentitySpec *identity.Spec `json:"identity_spec"`
	IdentityInfo *identity.Info `json:"identity_info"`
}

func (n *NodeSelectIdentityEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeSelectIdentityEnd) UserIdentity() *identity.Info {
	return n.IdentityInfo
}

func (n *NodeSelectIdentityEnd) UserID() string {
	if n.IdentityInfo == nil {
		return ""
	}
	return n.IdentityInfo.UserID
}
