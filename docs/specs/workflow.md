# Workflow

This document specifies the design of workflow.

A workflow represents a user interaction that involves multiple steps.
Examples of workflows are authentication, adding a new identity, verifying an email address, etc.

## Intent

The core of a workflow is its intent.
The intent controls how the workflow proceeds.
The kind of an intent and its JSON schema is part of the public API.
Some intent cannot be instantiated by the public API.

```golang
type Intent interface {
	Kind() string
	JSONSchema() *validation.SimpleSchema
}

type Workflow struct {
	Intent Intent
}
```

## Node

Each workflow has a series of nodes.
Each node is either a simple node, or a sub-workflow node.
A workflow proceeds by appending a new node at the end of its node list.
The kind of a node and its output data is part of the public API.

```golang
type NodeType string

const (
	NodeTypeSimple      NodeType = "SIMPLE"
	NodeTypeSubWorkflow NodeType = "SUB_WORKFLOW"
)

type NodeSimple interface{
	Kind() string
}

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

## Input

Both intent and node can react to inputs.
Reaction to inputs can result in transition to a new node, or update the node.
The kind of an input and its JSON schema is part of the public API.
Some input cannot be instantiated by the public API.

```golang
var ErrEOF = errors.New("eof")

type Input interface {
	Kind() string
	JSONSchema() *validation.SimpleSchema
}

type Intent interface {
	Kind() string
	JSONSchema() *validation.SimpleSchema
	CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error)
	ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error)
}

type NodeSimple interface {
	Kind() string
	CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error)
	ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error)
}
```

If the workflow has at least one node, the last node react to inputs first.

If the node can react to some inputs, the inputs are used.
If the node returns ErrEOF, the intent is asked to react to inputs.
If the node returns other error, the error is returned.
If the node can react to no input without error, it is a programming error.

If the intent can react to some inputs, the inputs are used.
If the intent returns ErrEOF, the workflow is finished.
If the intent return other error, the error is returned.
If the intent can react to no input without error, it is a programming error.

## Effects

Both intent and node can have effects attached to them.
There are 2 kinds of effects, namely run-effect and on-commit effects.

```golang
type Effect interface {}

type Intent interface {
	Kind() string
	JSONSchema() *validation.SimpleSchema
	CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error)
	ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error)
	GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) (effs []Effect, err error)
}

type NodeSimple interface {
	Kind() string
	CanReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow) ([]Input, error)
	ReactTo(ctx context.Context, deps *Dependencies, workflow *Workflow, input Input) (*Node, error)
	GetEffects(ctx context.Context, deps *Dependencies, workflow *Workflow) (effs []Effect, err error)
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
follow the rules outlined in [Input](#input) to find which node or intent to handle the input.

As a special case, the node can return ErrSameNode to perform an immediate side effect without adding a new node.

As a special case, the node can return ErrUpdateNode to update the node instead of adding a new node.

## HTTP API

The HTTP API is intended to be used by a web frontend to drive a workflow.

In a typical case, Authgear redirects to the frontend.
Based on the path, the frontend creates a suitable workflow.
The frontend is assumed to have a compatible UI to drive the workflow.
When the workflow is done, the frontend redirects to the given URI.

### Response

The response of HTTP API is as follows:

```json
{
  "action": {
    "type": "redirect",
    "redirect_uri": "https://myapp.authgearapps.com/oauth/authorize"
  },
  "workflow": {
    "workflow_id": "{workflow_id}",
    "instance_id": "{instance_id}",
    "intent": {
      "kind": "{intent_kind}",
      "data": {}
    },
    "nodes": [
      {
        "type": "SIMPLE",
        "simple": {
          "kind": "{node_kind}",
          "data": {}
        }
      },
      {
        "type": "SUB_WORKFLOW",
        "workflow": {
          "intent": {
            "kind": "{intent_kind}"
            "data": {}
          },
          "nodes": []
        }
      }
    ]
  }
}
```

- `action`: Tell the frontend what the action should it take.
- `action.type`:
  - When the value is `redirect`, the frontend must redirect the end-user to `action.redirect_uri`.
  - When the value is `continue`, the frontend must continue the workflow with its UI.
  - When the value is `finish`, `action.redirect_uri` may be present. If `action.redirect_uri` is present, the frontend must redirect the end-user there. Otherwise, the frontend must display a proper end screen.
- `workflow`: The workflow object.
- `workflow.workflow_id`: The workflow ID. It is the same across instances.
- `workflow.instance_id`: The instance ID. A workflow can have many instances. The frontend implements back-forward by keep tracking of instances.
- `workflow.intent`: The intent object.
- `workflow.nodes`: The list of nodes the instance of the workflow has gone through.

- `intent.kind`: The kind of the intent. The frontend must know the kind in order to create a new workflow.
- `intent.data`: The data of the intent. Vary by kind.

- `node.type`: The type of the node.
  - When the value is `SIMPLE`, then `node.simple` is present.
  - When the value is `SUB_WORKFLOW`, then `node.workflow` is present.
- `node.simple`: The simple node object.
- `node.simple.kind`: The kind of the simple node.
- `node.simple.data`: The data of the simple node.
- `node.workflow`: The sub-workflow object.
- `node.workflow.intent`: The intent object.
- `node.workflow.nodes`: The list of nodes the sub-workflow has gone through.

### POST /api/workflow/v1/

Create a new workflow by specifying an intent.

Request body

```json
{
  "intent": {
    "kind": "{intent_kind}",
    "data": {}
  }
}
```

### GET /api/workflow/v1/{instance-id}

Retrieve a workflow by the instance ID.

### POST /api/workflow/v1/{instance-id}

Feed an input to the workflow to drive it.

Request body

```json
{
  "input": {
    "kind": "{input_kind}",
    "data": {}
  }
}
```

> There is a choice between support multiple input in a single request, and support single input in a request.
> The later one makes the HTTP API implementation easier so it was chosen.
> If the frontend collects information more than the current node is requiring, it has to feed the input one by one in correct order.

### Redirect to an external OAuth provider

> This section is provisional.

Typically the end-user stays in the frontend for the entire workflow.
However, if signing in with an external OAuth provider is configured,
the frontend must redirect to the external OAuth provider in the middle of the workflow.
The frontend must be prepared for being redirected again to resume the workflow.

The frontend feeds an input to select a provider.

```json
{
  "input": {
    "kind": "SelectOAuthProvider",
    "data": {
      "alias": "google",
      "redirect_uri": "https://myapp.com/oauth-redirect"
    }
  }
}
```

- `input.data.alias`: Indicate the OAuth provider the end-user has chosen.
- `input.data.redirect_uri`: The redirect URI to the frontend. The URI must have the same origin as the frontend.

And then the frontend will receive a response to telling it to redirect.

```json
{
  "action": {
    "type": "redirect",
    "redirect_uri": "https://accounts.google.com/o/oauth2/v2/auth"
  },
  "workflow": {}
}
```

Finally the frontend will be visited again with `https://myapp.com/oauth-redirect`.
The frontend can include necessary information in the redirect URI to make the resumption possible.
