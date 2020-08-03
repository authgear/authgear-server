package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeDoCreateIdentity{})
}

type EdgeDoCreateIdentity struct {
	Identity *identity.Info
}

func (e *EdgeDoCreateIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeDoCreateIdentity{
		Identity: e.Identity,
	}, nil
}

type NodeDoCreateIdentity struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeDoCreateIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		if _, err := ctx.Identities.CheckDuplicated(n.Identity); err != nil {
			if errors.Is(err, identity.ErrIdentityAlreadyExists) {
				return newinteraction.ErrDuplicatedIdentity
			}
			return err
		}
		if err := ctx.Identities.Create(n.Identity); err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	// TODO(interaction): dispatch identity creation event if not creating user

	return nil
}

func (n *NodeDoCreateIdentity) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeDoCreateIdentity) UserIdentity() *identity.Info {
	return n.Identity
}

func (n *NodeDoCreateIdentity) UserNewIdentity() *identity.Info {
	return n.Identity
}
