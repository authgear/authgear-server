package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/lib/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeRemoveIdentity{})
}

type InputRemoveIdentity interface {
	GetIdentityType() authn.IdentityType
	GetIdentityID() string
}

type EdgeRemoveIdentity struct{}

func (e *EdgeRemoveIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputRemoveIdentity)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	identityType := input.GetIdentityType()
	identityID := input.GetIdentityID()

	info, err := ctx.Identities.Get(userID, identityType, identityID)
	if err != nil {
		return nil, err
	}

	return &NodeRemoveIdentity{
		IdentityInfo: info,
	}, nil
}

type NodeRemoveIdentity struct {
	IdentityInfo *identity.Info `json:"identity_info"`
}

func (n *NodeRemoveIdentity) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeRemoveIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeRemoveIdentity) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
