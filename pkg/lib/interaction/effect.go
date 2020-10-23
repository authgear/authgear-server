package interaction

type Effect interface {
	// apply the effect with graph when the current node is at nodeIndex.
	apply(ctx *Context, graph *Graph, nodeIndex int) error
}

type EffectRun func(ctx *Context, graph *Graph, nodeIndex int) error

func (e EffectRun) apply(ctx *Context, graph *Graph, nodeIndex int) error {
	if ctx.IsCommitting {
		return nil
	}
	slicedGraph := *graph
	slicedGraph.Nodes = slicedGraph.Nodes[:nodeIndex+1]
	return e(ctx, &slicedGraph, nodeIndex)
}

type EffectOnCommit func(ctx *Context, graph *Graph, nodeIndex int) error

func (e EffectOnCommit) apply(ctx *Context, graph *Graph, nodeIndex int) error {
	if !ctx.IsCommitting {
		return nil
	}
	return e(ctx, graph, nodeIndex)
}
