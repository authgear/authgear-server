package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeDoRemoveIdentity{})
}

type EdgeDoRemoveIdentity struct {
	Identity *identity.Info
}

func (e *EdgeDoRemoveIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoRemoveIdentity{
		Identity: e.Identity,
	}, nil
}

type NodeDoRemoveIdentity struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeDoRemoveIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		userID := graph.MustGetUserID()
		identityInfos, err := ctx.Identities.ListByUser(userID)
		if err != nil {
			return err
		}

		if len(identityInfos) <= 1 {
			return newinteraction.ErrCannotRemoveLastIdentity
		}

		err = ctx.Identities.Delete(n.Identity)
		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoRemoveIdentity) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}
