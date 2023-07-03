package workflowconfig

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
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

func (*IntentSignupFlow) Kind() string {
	return "workflowconfig.IntentSignupFlow"
}

func (*IntentSignupFlow) JSONSchema() *validation.SimpleSchema {
	return IntentSignupSchema
}

func (i *IntentSignupFlow) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	f, err := FindSignupFlow(deps.Config.Workflow, i.SignupFlow)
	if err != nil {
		return nil, err
	}

	// The list of nodes looks like
	// 1 NodeDoCreateUser
	// N Nodes for each N steps
	// 1 IntentCreateSession
	// So at the end of the flow, it will have 2+N nodes.
	if len(w.Nodes) >= 2+len(f.Steps) {
		return nil, workflow.ErrEOF
	}

	return nil, nil
}

func (i *IntentSignupFlow) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	f, err := FindSignupFlow(deps.Config.Workflow, i.SignupFlow)
	if err != nil {
		return nil, err
	}

	switch {
	// Before all steps
	case len(w.Nodes) == 0:
		return workflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	// After all steps
	case len(w.Nodes) == 1+len(f.Steps):
		// FIXME(workflow): create session
		break
	// During the steps.
	default:
		// Offset the NodeDoCreateUser
		nextStepIndex := len(w.Nodes) - 1
		step := f.Steps[nextStepIndex]
		switch step.Type {
		case config.WorkflowSignupFlowStepTypeIdentify:
			// FIXME(workflow): create identity
		case config.WorkflowSignupFlowStepTypeVerify:
			// FIXME(workflow): verify claim in the target step.
		case config.WorkflowSignupFlowStepTypeAuthenticate:
			// FIXME(workflow): create authenticator
		case config.WorkflowSignupFlowStepTypeUserProfile:
			// FIXME(workflow): fill user profile
		}
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentSignupFlow) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	// FIXME(workflow): perform signup effects.
	return nil, nil
}

func (*IntentSignupFlow) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
