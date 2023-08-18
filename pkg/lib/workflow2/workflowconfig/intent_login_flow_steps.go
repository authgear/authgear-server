package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterIntent(&IntentLoginFlowSteps{})
}

type IntentLoginFlowSteps struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
}

var _ workflow.Intent = &IntentLoginFlowSteps{}
var _ workflow.Milestone = &IntentLoginFlowSteps{}
var _ MilestoneNestedSteps = &IntentLoginFlowSteps{}

func (*IntentLoginFlowSteps) Kind() string {
	return "workflowconfig.IntentLoginFlowSteps"
}

func (*IntentLoginFlowSteps) Milestone()            {}
func (*IntentLoginFlowSteps) MilestoneNestedSteps() {}

func (i *IntentLoginFlowSteps) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
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

func (i *IntentLoginFlowSteps) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, _ workflow.Input) (*workflow.Node, error) {
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
		n, err := NewIntentLoginFlowStepAuthenticate(ctx, deps, workflows, &IntentLoginFlowStepAuthenticate{
			LoginFlow:   i.LoginFlow,
			StepID:      step.ID,
			JSONPointer: JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.userID(workflows),
		})
		if err != nil {
			return nil, err
		}

		return workflow.NewSubWorkflow(n), nil
	case config.WorkflowLoginFlowStepTypeChangePassword:
		return workflow.NewSubWorkflow(&IntentLoginFlowStepChangePassword{
			LoginFlow:   i.LoginFlow,
			StepID:      step.ID,
			JSONPointer: JSONPointerForStep(i.JSONPointer, nextStepIndex),
			UserID:      i.userID(workflows),
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentLoginFlowSteps) steps(o config.WorkflowObject) []config.WorkflowObject {
	steps, ok := o.GetSteps()
	if !ok {
		panic(fmt.Errorf("workflow: workflow object does not have steps %T", o))
	}

	return steps
}

func (*IntentLoginFlowSteps) userID(workflows workflow.Workflows) string {
	userID, err := getUserID(workflows)
	if err != nil {
		panic(err)
	}
	return userID
}
