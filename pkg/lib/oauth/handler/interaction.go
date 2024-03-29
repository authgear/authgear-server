package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type GraphService interface {
	NewGraph(ctx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	DryRun(contextValue interaction.ContextValues, fn func(*interaction.Context) (*interaction.Graph, error)) error
	Run(contextValue interaction.ContextValues, graph *interaction.Graph) error
	Accept(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (*interaction.Graph, []interaction.Edge, error)
}
