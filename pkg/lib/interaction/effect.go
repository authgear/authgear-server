package interaction

import (
	"context"
)

type Effect interface {
	// apply the effect with graph when the current node is at nodeIndex.
	apply(goCtx context.Context, ctx *Context, graph *Graph, nodeIndex int) error
}

type EffectRun func(goCtx context.Context, ctx *Context, graph *Graph, nodeIndex int) error

func (e EffectRun) apply(goCtx context.Context, ctx *Context, graph *Graph, nodeIndex int) error {
	if ctx.IsCommitting {
		return nil
	}
	slicedGraph := *graph
	slicedGraph.Nodes = slicedGraph.Nodes[:nodeIndex+1]
	return e(goCtx, ctx, &slicedGraph, nodeIndex)
}

type EffectOnCommit func(goCtx context.Context, ctx *Context, graph *Graph, nodeIndex int) error

func (e EffectOnCommit) apply(goCtx context.Context, ctx *Context, graph *Graph, nodeIndex int) error {
	if !ctx.IsCommitting {
		return nil
	}
	return e(goCtx, ctx, graph, nodeIndex)
}
