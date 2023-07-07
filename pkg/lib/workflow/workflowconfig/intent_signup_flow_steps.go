package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentSignupFlowSteps{})
}

var IntentSignupFlowStepsSchema = validation.NewSimpleSchema(`{}`)

type IntentSignupFlowSteps struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

func (*IntentSignupFlowSteps) Kind() string {
	return "workflowconfig.IntentSignupFlowSteps"
}

func (*IntentSignupFlowSteps) JSONSchema() *validation.SimpleSchema {
	return IntentSignupFlowStepsSchema
}

func (i *IntentSignupFlowSteps) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	current, err := i.current(deps, workflows.Nearest)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	if len(workflows.Nearest.Nodes) < len(steps) {
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowSteps) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := i.current(deps, workflows.Nearest)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(workflows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.WorkflowSignupFlowStep)

	switch step.Type {
	case config.WorkflowSignupFlowStepTypeIdentify:
		return workflow.NewSubWorkflow(&IntentSignupFlowIdentify{
			SignupFlow:  i.SignupFlow,
			StepID:      step.ID,
			JSONPointer: JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		}), nil
	case config.WorkflowSignupFlowStepTypeVerify:
		return workflow.NewSubWorkflow(&IntentSignupFlowVerify{
			SignupFlow:  i.SignupFlow,
			StepID:      step.ID,
			JSONPointer: JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.UserID,
		}), nil
	case config.WorkflowSignupFlowStepTypeAuthenticate:
		// FIXME(workflow): create authenticator
	case config.WorkflowSignupFlowStepTypeUserProfile:
		// FIXME(workflow): fill user profile
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentSignupFlowSteps) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentSignupFlowSteps) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentSignupFlowSteps) current(deps *workflow.Dependencies, w *workflow.Workflow) (config.WorkflowObject, error) {
	root, err := findSignupFlow(deps.Config.Workflow, i.SignupFlow)
	if err != nil {
		return nil, err
	}

	entries, err := Traverse(root, i.JSONPointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func (i *IntentSignupFlowSteps) steps(o config.WorkflowObject) []config.WorkflowObject {
	steps, ok := o.GetSteps()
	if !ok {
		panic(fmt.Errorf("workflow: workflow object does not have steps %T", o))
	}

	return steps
}
