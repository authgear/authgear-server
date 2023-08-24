# Workflow

> Warning: Workflow was an experiment and it was abandoned.
> The replacement of it is [Authentication Flow](./authentication-flow.md).

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

### CanReactTo and ReactTo

The return value of CanReactTo has the following meanings:

- `(nil, nil)`: The node or the intent can transition to a new node without input. ReactTo of the node or the intent will be called.
- `([ ... ], nil)`: The node or the intent can react to the returned inputs. ReactTo of the node or the intent will be called.
- `(nil, ErrEOF)` by a node: The node cannot react to input. The CanReactTo of the intent is called instead.
- `(nil, ErrEOF)` by an intent: The workflow is finished.
- `(nil, error)`: Some error occurred.

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

## HTTP API v1

The HTTP API v1 is intended to be used by a web frontend to drive a workflow.

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

### POST /api/v1/workflows/

Create a new workflow by specifying an intent.

Please read the [user agent binding section](#user-agent-binding) for details of the `bind_user_agent` flag.

Request body

```json
{
  "intent": {
    "kind": "{intent_kind}",
    "data": {}
  },
  "bind_user_agent": true
}
```

### GET /api/v1/workflows/{workflow-id}/instances/{instance-id}

Retrieve a workflow by the instance ID.

### POST /api/v1/workflows/{workflow-id}/instances/{instance-id}

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

### GET /api/v1/workflows/{workflow-id}/ws

Subscribe to workflow events.

Clients should not send any messages to server.
Server would send following messages through WebSocket:

#### Refresh

Indicates out-of-band data (e.g. OTP code) of the workflow has changed.
Clients should re-fetch the current instance of workflow to get the latest state.

```json
{ "kind": "refresh" }
```

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

### Intent for account migration (Experimental)

The Intent for account migration allows the developer to create a workflow for migrating an account.

[![](https://mermaid.ink/img/pako:eNqdkltLw0AQhf_KsM9V3_NQKK1gES9YxQfjw7iZNEuT3bg7ayml_92JSXohUMVAIJk558x-u7tV2mWkEhXoM5LVNDO49FilFuSp0bPRpkbLMIlcLAn9sDONgV31Mh92bpxbAQaYkXVwBa_0UUil1fWui_G4j05g6gmZwFgmy3dGVsI00dpFCVs7v8pLt27tvUfsfVICT8TR270S5rORZAVGAZOfM4MfvdNEGXBBB_vacAHPD7fX94OZDZm4MIRWAOxa2i-DcIBsSnv11MlSfNQMb-8mE0DDm8sOclGTBrSZtFCmNE2N7PxJP_euOl5Pl36g6Pj_Fz9g_MvuNJUg5vD7qZw7iGNxiBWdTGrUzatGqiJfocnkvm5_qkp0FaUqkc-Mcowlpyq1O5EKp1tsrFaJ7DmNVKwzAe2ut0pyLAPtvgFMVAp7?type=png)](https://mermaid.live/edit#pako:eNqdkltLw0AQhf_KsM9V3_NQKK1gES9YxQfjw7iZNEuT3bg7ayml_92JSXohUMVAIJk558x-u7tV2mWkEhXoM5LVNDO49FilFuSp0bPRpkbLMIlcLAn9sDONgV31Mh92bpxbAQaYkXVwBa_0UUil1fWui_G4j05g6gmZwFgmy3dGVsI00dpFCVs7v8pLt27tvUfsfVICT8TR270S5rORZAVGAZOfM4MfvdNEGXBBB_vacAHPD7fX94OZDZm4MIRWAOxa2i-DcIBsSnv11MlSfNQMb-8mE0DDm8sOclGTBrSZtFCmNE2N7PxJP_euOl5Pl36g6Pj_Fz9g_MvuNJUg5vD7qZw7iGNxiBWdTGrUzatGqiJfocnkvm5_qkp0FaUqkc-Mcowlpyq1O5EKp1tsrFaJ7DmNVKwzAe2ut0pyLAPtvgFMVAp7)

## HTTP API v2

HTTP API v1 puts variable in path, and accept query parameters.
This causes Access-Control-Max-Age to has no actual effect.

### POST /api/v2/workflows

Create a new workflow by specifying an intent.

Please read the [user agent binding section](#user-agent-binding) for details of the `bind_user_agent` flag.

Request body

```json
{
  "action": "create",
  "url_query": "",
  "intent": {
    "kind": "{intent_kind}",
    "data": {}
  },
  "bind_user_agent": true
}
```

Feed an input to the workflow to drive it.

Request body

```json
{
  "action": "input",
  "workflow_id": "",
  "instance_id": "",
  "input": {
    "kind": "{input_kind}",
    "data": {}
  }
}
```

Feed more than 1 input to the workflow.

Request body

```json
{
  "action": "batch_input",
  "workflow_id": "",
  "instance_id": "",
  "batch_input": [
    {
      "kind": "{input_kind}",
      "data": {}
    },
    {
      "kind": "{input_kind}",
      "data": {}
    }
  ]
}
```

Retrieve a workflow by the instance ID

Request body

```json
{
  "action": "get",
  "workflow_id": "",
  "instance_id": ""
}
```

#### Configuration in `authgear.yaml`

```yaml
account_migration:
  hook:
    url: authgeardeno:///deno/migration.ts
    timeout: 10
```

The developer should send the migration token (generated by the developer) to the workflow. Authgear will send the migration token to the hook, and the hook should return the identities' and authenticators' MigrateSpec.

#### Input of IntentMigrateAccount

```json
{
  "input": {
    "kind": "latte.InputTakeMigrationToken",
    "data": {
      "migration_token": "TOKEN"
    }
  }
}
```

#### Hook request

```json
{
  "migration_token": "TOKEN"
}
```

#### Hook response

```json
{
  "identities": [
    {
      "type": "login_id",
      "login_id": {
        "key": "email",
        "type": "email",
        "value": "test@example.com"
      }
    }
  ],
  "authenticators": [
    {
      "type": "oobotp",
      "oobotp": {
        "email": "test@example.com"
      }
    }
  ]
}
```

The developer has the responsibility to ensure the identities and authenticators combination is valid.
Otherwise, the user won't be able to login with the identity. e.g. The login id identity should have its primary authenticator.

Only login id identity and OOB OTP authenticator are supported in this stage.

#### The identity spec

```go

type MigrateSpec struct {
	Type model.IdentityType `json:"type"`

	LoginID *MigrateLoginIDSpec `json:"login_id,omitempty"`
}

type MigrateLoginIDSpec struct {
	Key   string               `json:"key"`
	Type  model.LoginIDKeyType `json:"type"`
	Value string               `json:"value"`
}
```

### The authenticator spec

```go
type MigrateSpec struct {
	Type model.AuthenticatorType `json:"type,omitempty"`

	OOBOTP *MigrateOOBOTPSpec `json:"oobotp,omitempty"`
}

type MigrateOOBOTPSpec struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}
```

## User Agent Binding

By default, only the user agent which created the workflow is able to use the workflow in workflow apis. This is done by setting a token into cookies of the user agent which creates the workflow, and check this token in any subsequent api calls.

User agent binding can be disable by setting `bind_user_agent` to `false` when creating the workflow.
