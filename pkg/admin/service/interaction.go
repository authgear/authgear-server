package service

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type InteractionGraphService interface {
	NewGraph(ctx context.Context, interactionCtx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	DryRun(ctx context.Context, contextValue interaction.ContextValues, fn func(ctx context.Context, interactionCtx *interaction.Context) (*interaction.Graph, error)) error
	Run(ctx context.Context, contextValue interaction.ContextValues, graph *interaction.Graph) error
	Accept(ctx context.Context, interactionCtx *interaction.Context, graph *interaction.Graph, input interface{}) (*interaction.Graph, []interaction.Edge, error)
}

type InteractionService struct {
	Graph InteractionGraphService
}

func (s *InteractionService) Perform(ctx context.Context, intent interaction.Intent, input interface{}) (*interaction.Graph, error) {
	var graph *interaction.Graph
	err := s.Graph.DryRun(ctx, interaction.ContextValues{}, func(ctx context.Context, interactionCtx *interaction.Context) (*interaction.Graph, error) {
		var err error
		graph, err = s.Graph.NewGraph(ctx, interactionCtx, intent)
		if err != nil {
			return nil, err
		}

		graph, _, err = s.Graph.Accept(ctx, interactionCtx, graph, input)
		if err != nil {
			return nil, err
		}

		return graph, nil
	})
	if err != nil {
		return graph, err
	}

	err = s.Graph.Run(ctx, interaction.ContextValues{}, graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
