# Workflow

This document specifies the design of workflow.

A workflow represents a user interaction that involves multiple steps.
Examples of workflows are authentication, adding a new identity, verifying an email address, etc.

## Intent

The core of a workflow is its intent.
The intent controls how the workflow proceeds.

```golang
type Intent interface {}

type Workflow struct {
    Intent Intent
}
```

## Nodes

Each workflow has a series of nodes.
Each node is either a simple node, or a sub-workflow node.
A workflow proceeds by appending a new node at the end of its node list.

```golang
type NodeType string

const (
	NodeTypeSimple      NodeType = "SIMPLE"
	NodeTypeSubWorkflow NodeType = "SUB_WORKFLOW"
)

type NodeSimple interface{}

type Node struct {
    Type        NodeType
    // When Type is NodeTypeSimple, Simple is non-nil.
    Simple      NodeSimple
    // When Type is NodeTypeSubWorkflow, SubWorkflow is non-nil.
    SubWorkflow *Workflow
}

type Workflow struct {
    Intent Intent
    Nodes  []Node
}
```

## Edges

Both intent and node can derive edges.
Edges react to input.
Edges can instantiate a new node, or update the node.
Input is any Go value satisfying a particular interface required by a specific edge.

```golang
var ErrEOF = errors.New("eof")

type Edge interface {
    // The edge tries casting input to a specific interface.
    Instantiate(ctx *Context, workflow *Workflow, input interface{}) (*Node, error)
}

type Intent interface {
	DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error)
}

type NodeSimple interface {
	DeriveEdges(ctx *Context) ([]Edge, error)
}
```

If the workflow has at least one node, the last node derives edges first.

If the node derives non-empty edges, the edges are used.
If the node returns ErrEOF, the intent is asked to derive edges.
If the node returns other error, the error is returned.
If the node derives empty edges without error, it is a programming error.

If the intent derives non-empty edges, the edges are used.
If the intent returns ErrEOF, the workflow is finished.
If the intent returns other error, the error is returned.
If the intent derives empty edges without error, it is a programming error.

## Effects

Both intent and node can have effects attached to them.
There are 2 kinds of effects, namely run-effect and on-commit effects.

```golang
type Intent interface {
	DeriveEdges(ctx *Context, workflow *Workflow) ([]Edge, error)
	GetEffects(ctx *Context, workflow *Workflow) (effs []Effect, err error)
}

type NodeSimple interface {
	DeriveEdges(ctx *Context) ([]Edge, error)
	GetEffects(ctx *Context) (effs []Effect, err error)
}
```

### Application of effects

When a workflow is restored from the database, the run-effects are applied.
When a node is appended to the workflow, only the effects of the node is applied.

### Rule 1 of effects: A run-effect is always applied.
A typical run-effect is mutation on the database.
The run-effect of a previous node is visible to a latter node.

### Rule 2 of effects: A on-commit effect is applied only when the **WHOLE** workflow is finished.
A typical on-commit effect is delivering the events.

### Rule 3 of effects: All run-effects are applied **BEFORE** all on-commit effects.

### Rule 4 of effects: The effects of the nodes are applied **BEFORE** the effects of the intent.

### Rule 5 of effects: Intent **MUST** only have on-commit effects.
This restriction makes the reasoning of the order of application of effects much easier.
Imagine if an intent is allowed to have run-effects,
the workflow is restored, the run-effect of the intent is applied.
And then a new node is appended, the run-effect of the new node is applied.
The run-effect of the new node is applied **AFTER** the run-effect of the intent.
This breaks Rule 4.

## Accepting input
When input is fed to the workflow,
edges are derived by the workflow.
Each edge has a chance to react to the input.
If the edge does not react to the input, it must return ErrIncompatibleInput.

As a special case, the edge can return ErrSameNode to perform an immediate side effect without adding a new node.

As a special case, the edge can return ErrUpdateNode to update the node instead of adding a new node.
