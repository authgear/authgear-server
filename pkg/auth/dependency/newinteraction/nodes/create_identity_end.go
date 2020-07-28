package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityEnd{})
}

type EdgeCreateIdentityEnd struct {
	NewIdentity *identity.Info
}

func (e *EdgeCreateIdentityEnd) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	return &NodeCreateIdentityEnd{NewIdentity: e.NewIdentity}, nil
}

type NodeCreateIdentityEnd struct {
	NewIdentity *identity.Info `json:"new_identity"`
}

func (n *NodeCreateIdentityEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectRun(func(ctx *newinteraction.Context) error {
		if err := ctx.Identities.Validate(graph.GetUserNewIdentities()); err != nil {
			return err
		}

		if err := ctx.Identities.CheckDuplicated(n.NewIdentity); err != nil {
			if errors.Is(err, identity.ErrIdentityAlreadyExists) {
				return newinteraction.ErrDuplicatedIdentity
			}
			return err
		}
		if err := ctx.Identities.Create(n.NewIdentity); err != nil {
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

func (n *NodeCreateIdentityEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}

func (n *NodeCreateIdentityEnd) UserIdentity() *identity.Info {
	return n.NewIdentity
}

func (n *NodeCreateIdentityEnd) UserNewIdentity() *identity.Info {
	return n.NewIdentity
}
