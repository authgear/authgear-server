package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentLoginFlowStepIdentify{})
}

var IntentLoginFlowStepIdentifySchema = validation.NewSimpleSchema(`{}`)

type IntentLoginFlowStepIdentify struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
}

var _ WorkflowStep = &IntentLoginFlowStepIdentify{}

func (i *IntentLoginFlowStepIdentify) GetID() string {
	return i.StepID
}

func (i *IntentLoginFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentLoginFlowStepAuthenticateTarget = &IntentLoginFlowStepIdentify{}

func (*IntentLoginFlowStepIdentify) GetIdentity(_ context.Context, _ *workflow.Dependencies, workflows workflow.Workflows) *identity.Info {
	m, ok := FindMilestone[MilestoneDoUseIdentity](workflows.Nearest)
	if !ok {
		panic(fmt.Errorf("MilestoneDoUseIdentity is absent in IntentLoginFlowStepIdentify"))
	}

	info, ok := m.MilestoneDoUseIdentity()
	if !ok {
		panic(fmt.Errorf("MilestoneDoUseIdentity does not return identity"))
	}

	return info
}

var _ workflow.Intent = &IntentLoginFlowStepIdentify{}

func (*IntentLoginFlowStepIdentify) Kind() string {
	return "workflowconfig.IntentLoginFlowStepIdentify"
}

func (*IntentLoginFlowStepIdentify) JSONSchema() *validation.SimpleSchema {
	return IntentLoginFlowStepIdentifySchema
}

func (*IntentLoginFlowStepIdentify) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Let the input to select which identification method to use.
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeIdentificationMethod{},
		}, nil
	}

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoUseIdentity:
			// Handle nested steps.
			return nil, nil
		}
	}

	return nil, workflow.ErrEOF
}

func (i *IntentLoginFlowStepIdentify) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(workflows.Nearest.Nodes) == 0 {
		var inputTakeIdentificationMethod inputTakeIdentificationMethod
		if workflow.AsInput(input, &inputTakeIdentificationMethod) &&
			inputTakeIdentificationMethod.GetJSONPointer().String() == i.JSONPointer.String() {

			identification := inputTakeIdentificationMethod.GetIdentificationMethod()
			err = i.checkIdentificationMethod(step, identification)
			if err != nil {
				return nil, err
			}

			switch identification {
			case config.WorkflowIdentificationMethodEmail:
				fallthrough
			case config.WorkflowIdentificationMethodPhone:
				fallthrough
			case config.WorkflowIdentificationMethodUsername:
				return workflow.NewNodeSimple(&NodeUseIdentityLoginID{
					Identification: identification,
				}), nil
			case config.WorkflowIdentificationMethodOAuth:
				// FIXME(workflow): handle oauth
			case config.WorkflowIdentificationMethodPasskey:
				// FIXME(workflow): handle passkey
			case config.WorkflowIdentificationMethodSiwe:
				// FIXME(workflow): handle siwe
			}
		}
		return nil, workflow.ErrIncompatibleInput
	}

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoUseIdentity:
			identification := i.identificationMethod(workflows.Nearest)
			return workflow.NewSubWorkflow(&IntentLoginFlowSteps{
				LoginFlow:   i.LoginFlow,
				JSONPointer: i.jsonPointer(step, identification),
			}), nil
		}
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentLoginFlowStepIdentify) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (i *IntentLoginFlowStepIdentify) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return map[string]interface{}{
		"json_pointer": i.JSONPointer.String(),
	}, nil
}

func (*IntentLoginFlowStepIdentify) step(o config.WorkflowObject) *config.WorkflowLoginFlowStep {
	step, ok := o.(*config.WorkflowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (*IntentLoginFlowStepIdentify) checkIdentificationMethod(step *config.WorkflowLoginFlowStep, im config.WorkflowIdentificationMethod) error {
	var allAllowed []config.WorkflowIdentificationMethod

	for _, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Identification)
	}

	for _, allowed := range allAllowed {
		if im == allowed {
			return nil
		}
	}

	return InvalidIdentificationMethod.NewWithInfo("invalid identification method", apierrors.Details{
		"expected": allAllowed,
		"actual":   im,
	})
}

func (*IntentLoginFlowStepIdentify) identificationMethod(w *workflow.Workflow) config.WorkflowIdentificationMethod {
	if len(w.Nodes) == 0 {
		panic(fmt.Errorf("workflow: identification method not yet selected"))
	}

	switch n := w.Nodes[0].Simple.(type) {
	case *NodeUseIdentityLoginID:
		return n.Identification
	default:
		panic(fmt.Errorf("workflow: unexpected node: %T", w.Nodes[0].Simple))
	}
}

func (i *IntentLoginFlowStepIdentify) jsonPointer(step *config.WorkflowLoginFlowStep, im config.WorkflowIdentificationMethod) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("workflow: selected identification method is not allowed"))
}
