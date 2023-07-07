package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentSignupFlow{})
}

var IntentSignupSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["signup_flow"],
	"properties": {
		"signup_flow": { "type": "string" }
	}
}
`)

type IntentSignupFlow struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ workflow.Intent = &IntentSignupFlow{}

func (*IntentSignupFlow) Kind() string {
	return "workflowconfig.IntentSignupFlow"
}

func (*IntentSignupFlow) JSONSchema() *validation.SimpleSchema {
	return IntentSignupSchema
}

func (i *IntentSignupFlow) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// The list of nodes looks like
	// 1 NodeDoCreateUser
	// 1 IntentSignupFlowSteps
	// 1 IntentCreateSession
	// So at the end of the flow, it will have 3 nodes.
	if len(workflows.Nearest.Nodes) == 3 {
		return nil, workflow.ErrEOF
	}

	return nil, nil
}

func (i *IntentSignupFlow) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch {
	case len(workflows.Nearest.Nodes) == 0:
		return workflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	case len(workflows.Nearest.Nodes) == 1:
		return workflow.NewSubWorkflow(&IntentSignupFlowSteps{
			SignupFlow:  i.SignupFlow,
			JSONPointer: i.JSONPointer,
			UserID:      i.userID(workflows.Nearest),
		}), nil
	case len(workflows.Nearest.Nodes) == 2:
		// FIXME(workflow): create session
		break
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentSignupFlow) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	// FIXME(workflow): perform signup effects.
	return nil, nil
}

func (*IntentSignupFlow) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentSignupFlow) userID(w *workflow.Workflow) string {
	node, ok := workflow.FindSingleNode[*NodeDoCreateUser](w)
	if !ok {
		panic(fmt.Errorf("expected userID"))
	}
	return node.UserID
}
