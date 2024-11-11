package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeConfirmTerminateOtherSessionsEnd{})
}

type InputConfirmTerminateOtherSessionsEnd interface {
	GetIsConfirmed() bool
}

type EdgeConfirmTerminateOtherSessionsEnd struct {
}

func (e *EdgeConfirmTerminateOtherSessionsEnd) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	clientID := ctx.Request.URL.Query().Get("client_id")
	client := ctx.OAuthClientResolver.ResolveClient(clientID)
	if client == nil || client.MaxConcurrentSession != 1 {
		return &NodeConfirmTerminateOtherSessionsEnd{}, nil
	}
	existingGrants, err := ctx.OfflineGrants.ListClientOfflineGrants(goCtx, client.ClientID, graph.MustGetUserID())
	if err != nil {
		return nil, err
	}
	if len(existingGrants) == 0 {
		return &NodeConfirmTerminateOtherSessionsEnd{}, nil
	}

	var confirmInput InputConfirmTerminateOtherSessionsEnd
	if !interaction.Input(rawInput, &confirmInput) {
		return nil, interaction.ErrIncompatibleInput
	}

	return &NodeConfirmTerminateOtherSessionsEnd{}, nil
}

type NodeConfirmTerminateOtherSessionsEnd struct {
}

func (n *NodeConfirmTerminateOtherSessionsEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeConfirmTerminateOtherSessionsEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeConfirmTerminateOtherSessionsEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
