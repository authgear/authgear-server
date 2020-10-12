package service

import "github.com/authgear/authgear-server/pkg/lib/interaction"

type InteractionGraphService interface {
	NewGraph(ctx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	DryRun(webStateID string, fn func(*interaction.Context) (*interaction.Graph, error)) error
	Run(webStateID string, graph *interaction.Graph) error
}

type InteractionService struct {
	Graph InteractionGraphService
}

func (s *InteractionService) Perform(intent interaction.Intent, input interface{}) (*interaction.Graph, error) {
	// TODO(admin): how to assign a state ID?
	stateID := ""
	var graph *interaction.Graph
	err := s.Graph.DryRun(stateID, func(ctx *interaction.Context) (*interaction.Graph, error) {
		var err error
		graph, err = s.Graph.NewGraph(ctx, intent)
		if err != nil {
			return nil, err
		}

		graph, _, err = graph.Accept(ctx, input)
		if err != nil {
			return nil, err
		}

		return graph, nil
	})
	if err != nil {
		return graph, err
	}

	err = s.Graph.Run(stateID, graph)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
