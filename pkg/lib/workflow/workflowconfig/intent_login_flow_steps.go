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
	workflow.RegisterPrivateIntent(&IntentLoginFlowSteps{})
}

var IntentLoginFlowStepsSchema = validation.NewSimpleSchema(`{}`)

type IntentLoginFlowSteps struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

func (*IntentLoginFlowSteps) Kind() string {
	return "workflow.IntentLoginFlowSteps"
}

func (*IntentLoginFlowSteps) JSONSchema() *validation.SimpleSchema {
	return IntentLoginFlowStepsSchema
}

func (i *IntentLoginFlowSteps) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	if len(workflows.Nearest.Nodes) < len(steps) {
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentLoginFlowSteps) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}

	steps := i.steps(current)
	nextStepIndex := len(workflows.Nearest.Nodes)
	step := steps[nextStepIndex].(*config.WorkflowLoginFlowStep)

	switch step.Type {
	case config.WorkflowLoginFlowStepTypeIdentify:
		return workflow.NewSubWorkflow(&IntentLoginFlowStepIdentify{
			LoginFlow:   i.LoginFlow,
			StepID:      step.ID,
			JSONPointer: JSONPointerForStep(i.JSONPointer, nextStepIndex),
		}), nil
	case config.WorkflowLoginFlowStepTypeAuthenticate:
		// FIXME
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentLoginFlowSteps) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*IntentLoginFlowSteps) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentLoginFlowSteps) steps(o config.WorkflowObject) []config.WorkflowObject {
	steps, ok := o.GetSteps()
	if !ok {
		panic(fmt.Errorf("workflow: workflow object does not have steps %T", o))
	}

	return steps
}
