package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepIdentify{})
}

type IntentLoginFlowStepIdentify struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepID        string                 `json:"step_id,omitempty"`
}

var _ FlowStep = &IntentLoginFlowStepIdentify{}

func (i *IntentLoginFlowStepIdentify) GetID() string {
	return i.StepID
}

func (i *IntentLoginFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ IntentLoginFlowStepAuthenticateTarget = &IntentLoginFlowStepIdentify{}

func (*IntentLoginFlowStepIdentify) GetIdentity(_ context.Context, _ *authflow.Dependencies, flows authflow.Flows) *identity.Info {
	m, ok := authflow.FindMilestone[MilestoneDoUseIdentity](flows.Nearest)
	if !ok {
		panic(fmt.Errorf("MilestoneDoUseIdentity is absent in IntentLoginFlowStepIdentify"))
	}

	info := m.MilestoneDoUseIdentity()
	return info
}

var _ authflow.Intent = &IntentLoginFlowStepIdentify{}

func (*IntentLoginFlowStepIdentify) Kind() string {
	return "IntentLoginFlowStepIdentify"
}

func (i *IntentLoginFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	current, err := flowObject(deps, i.FlowReference, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		return &InputSchemaLoginFlowStepIdentify{
			OneOf: step.OneOf,
		}, nil
	}

	_, identityUsed := authflow.FindMilestone[MilestoneDoUseIdentity](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityUsed && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentLoginFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := flowObject(deps, i.FlowReference, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeIdentificationMethod inputTakeIdentificationMethod
		if authflow.AsInput(input, &inputTakeIdentificationMethod) {

			identification := inputTakeIdentificationMethod.GetIdentificationMethod()

			switch identification {
			case config.AuthenticationFlowIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowIdentificationPhone:
				fallthrough
			case config.AuthenticationFlowIdentificationUsername:
				return authflow.NewNodeSimple(&NodeUseIdentityLoginID{
					Identification: identification,
				}), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, identityUsed := authflow.FindMilestone[MilestoneDoUseIdentity](flows.Nearest)
	_, nestedStepsHandled := authflow.FindMilestone[MilestoneNestedSteps](flows.Nearest)

	switch {
	case identityUsed && !nestedStepsHandled:
		identification := i.identificationMethod(flows.Nearest)
		return authflow.NewSubFlow(&IntentLoginFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, identification),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (*IntentLoginFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentLoginFlowStepIdentify) identificationMethod(w *authflow.Flow) config.AuthenticationFlowIdentification {
	m, ok := authflow.FindMilestone[MilestoneIdentificationMethod](w)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneIdentificationMethod()

	return im
}

func (i *IntentLoginFlowStepIdentify) jsonPointer(step *config.AuthenticationFlowLoginFlowStep, im config.AuthenticationFlowIdentification) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}
