package handler

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type GraphService interface {
	NewGraph(ctx context.Context, interactionCtx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	DryRun(ctx context.Context, contextValue interaction.ContextValues, fn func(ctx context.Context, interactionCtx *interaction.Context) (*interaction.Graph, error)) error
	Run(ctx context.Context, contextValue interaction.ContextValues, graph *interaction.Graph) error
	Accept(ctx context.Context, interactionCtx *interaction.Context, graph *interaction.Graph, input interface{}) (*interaction.Graph, []interaction.Edge, error)
}
