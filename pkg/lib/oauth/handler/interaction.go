package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type GraphService interface {
	NewGraph(ctx *interaction.Context, intent interaction.Intent) (*interaction.Graph, error)
	DryRun(webStateID string, fn func(*interaction.Context) (*interaction.Graph, error)) error
	Run(webStateID string, graph *interaction.Graph, preserveGraph bool) error
}
