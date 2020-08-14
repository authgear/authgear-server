package handler

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type GraphService interface {
	NewGraph(ctx *newinteraction.Context, intent newinteraction.Intent) (*newinteraction.Graph, error)
	DryRun(webStateID string, fn func(*newinteraction.Context) (*newinteraction.Graph, error)) error
	Run(webStateID string, graph *newinteraction.Graph, preserveGraph bool) error
}
