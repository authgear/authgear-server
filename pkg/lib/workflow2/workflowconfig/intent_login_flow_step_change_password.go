package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

type IntentLoginFlowStepChangePasswordTarget interface {
	GetPasswordAuthenticator(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (*authenticator.Info, bool)
}

func init() {
	workflow.RegisterIntent(&IntentLoginFlowStepChangePassword{})
}

type IntentLoginFlowStepChangePassword struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentLoginFlowStepChangePassword{}

func (i *IntentLoginFlowStepChangePassword) GetID() string {
	return i.StepID
}

func (i *IntentLoginFlowStepChangePassword) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ workflow.Intent = &IntentLoginFlowStepChangePassword{}
var _ workflow.Boundary = &IntentLoginFlowStepChangePassword{}

func (*IntentLoginFlowStepChangePassword) Kind() string {
	return "workflowconfig.IntentLoginFlowStepChangePassword"
}

func (i *IntentLoginFlowStepChangePassword) Boundary() string {
	return i.JSONPointer.String()
}

func (*IntentLoginFlowStepChangePassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Look up the password authenticator to change.
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentLoginFlowStepChangePassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}

	step := i.step(current)
	targetStepID := step.TargetStep

	targetStepWorkflow, err := FindTargetStep(workflows.Root, targetStepID)
	if err != nil {
		return nil, err
	}

	target, ok := targetStepWorkflow.Intent.(IntentLoginFlowStepChangePasswordTarget)
	if !ok {
		return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
			"target_step": targetStepID,
		})
	}

	info, ok := target.GetPasswordAuthenticator(ctx, deps, workflows.Replace(targetStepWorkflow))
	if !ok {
		// No need to change. End this workflow.
		return workflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	return workflow.NewNodeSimple(&NodeLoginFlowChangePassword{
		Authenticator: info,
	}), nil
}

func (*IntentLoginFlowStepChangePassword) step(o config.WorkflowObject) *config.WorkflowLoginFlowStep {
	step, ok := o.(*config.WorkflowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}
