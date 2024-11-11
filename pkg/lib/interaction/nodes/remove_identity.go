package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeRemoveIdentity{})
}

type InputRemoveIdentity interface {
	GetIdentityType() model.IdentityType
	GetIdentityID() string
}

type EdgeRemoveIdentity struct{}

func (e *EdgeRemoveIdentity) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputRemoveIdentity
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	identityID := input.GetIdentityID()

	info, err := ctx.Identities.Get(goCtx, identityID)
	if err != nil {
		return nil, err
	}

	if info.UserID != graph.MustGetUserID() {
		return nil, api.NewInvariantViolated(
			"IdentityNotBelongToUser",
			"identity does not belong to the user",
			nil,
		)
	}

	return &NodeRemoveIdentity{
		IdentityInfo: info,
	}, nil
}

type NodeRemoveIdentity struct {
	IdentityInfo *identity.Info `json:"identity_info"`
}

func (n *NodeRemoveIdentity) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeRemoveIdentity) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeRemoveIdentity) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
