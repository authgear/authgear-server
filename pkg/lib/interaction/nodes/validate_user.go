package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeValidateUser{})
}

type EdgeValidateUser struct {
}

func (e *EdgeValidateUser) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	u, err := ctx.Users.GetRaw(graph.MustGetUserID())
	if err != nil {
		return nil, err
	}

	var apiError *apierrors.APIError
	if err := u.CheckStatus(); err != nil {
		apiError = apierrors.AsAPIError(err)
	}

	return &NodeValidateUser{
		Error: apiError,
	}, nil
}

type NodeValidateUser struct {
	Error *apierrors.APIError `json:"error"`
}

func (n *NodeValidateUser) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeValidateUser) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeValidateUser) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
