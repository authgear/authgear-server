package handler

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type GraphService interface {
	NewGraph(ctx *newinteraction.Context, intent newinteraction.Intent) (*newinteraction.Graph, error)
	DryRun(fn func(*newinteraction.Context) (*newinteraction.Graph, error)) error
	Run(graph *newinteraction.Graph, preserveGraph bool) error
}
