package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

// EdgeIncompatibleInput always return ErrIncompatibleInput when instantiating
// to ensure graph won't end at unexpected node
type EdgeIncompatibleInput struct {
}

func (e *EdgeIncompatibleInput) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return nil, interaction.ErrIncompatibleInput
}

var _ interaction.Edge = &EdgeIncompatibleInput{}
