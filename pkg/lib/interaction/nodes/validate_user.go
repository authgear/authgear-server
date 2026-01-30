package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeValidateUser{})
}

type EdgeValidateUser struct {
}

func (e *EdgeValidateUser) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	u, err := ctx.Users.GetRaw(goCtx, graph.MustGetUserID())
	if err != nil {
		return nil, err
	}

	var apiError *apierrors.APIError
	now := ctx.Clock.NowUTC()
	if err := u.AccountStatus(now).Check(); err != nil {
		apiError = apierrors.AsAPIErrorWithContext(goCtx, err)
	}

	return &NodeValidateUser{
		Error: apiError,
	}, nil
}

type NodeValidateUser struct {
	Error *apierrors.APIError `json:"error"`
}

func (n *NodeValidateUser) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeValidateUser) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeValidateUser) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
