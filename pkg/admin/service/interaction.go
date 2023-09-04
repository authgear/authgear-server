package service

import "github.com/authgear/authgear-server/pkg/lib/interaction"

type InteractionGraphService interface {
	NewGraph(ctx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	DryRun(contextValues interaction.ContextValues, fn func(*interaction.Context) (*interaction.Graph, error)) error
	Run(contextValues interaction.ContextValues, graph *interaction.Graph) error
	Accept(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (*interaction.Graph, []interaction.Edge, error)
}

type InteractionService struct {
	Graph InteractionGraphService
}

func (s *InteractionService) Perform(intent interaction.Intent, input interface{}) (*interaction.Graph, error) {
	var graph *interaction.Graph
	err := s.Graph.DryRun(interaction.ContextValues{}, func(ctx *interaction.Context) (*interaction.Graph, error) {
		var err error
		graph, err = s.Graph.NewGraph(ctx, intent)
		if err != nil {
			return nil, err
		}

		graph, _, err = s.Graph.Accept(ctx, graph, input)
		if err != nil {
			return nil, err
		}

		return graph, nil
	})
	if err != nil {
		return graph, err
	}

	err = s.Graph.Run(interaction.ContextValues{}, graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
