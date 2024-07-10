package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentReauthFlowStepIdentify{})
}

type IntentReauthFlowStepIdentify struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	Options       []IdentificationOption `json:"options"`
}

var _ authflow.Intent = &IntentReauthFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentReauthFlowStepIdentify{}

func NewIntentReauthFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentReauthFlowStepIdentify) (*IntentReauthFlowStepIdentify, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options := []IdentificationOption{}
	for _, b := range step.OneOf {
		switch b.Identification {
		case config.AuthenticationFlowIdentificationIDToken:
			// Reauth flow identify step has no user interaction. It's called by our own backend to identify user. Thus, no bot protection is needed.
			var botProtection *config.AuthenticationFlowBotProtection = nil
			c := NewIdentificationOptionIDToken(b.Identification, botProtection, deps.Config.BotProtection)
			options = append(options, c)
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentReauthFlowStepIdentify) Kind() string {
	return "IntentReauthFlowStepIdentify"
}

func (i *IntentReauthFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, userIdentified := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseUser](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case len(flows.Nearest.Nodes) == 0 && authflow.GetIDToken(ctx) != "":
		// Special case: if id_token is available, use it automatically.
		return nil, nil
	case len(flows.Nearest.Nodes) == 0:
		// Let the input to select which identification method to use.
		flowRootObject, err := flowRootObject(deps, i.FlowReference)
		if err != nil {
			return nil, err
		}
		shouldBypassBotProtection := ShouldExistingResultBypassBotProtectionRequirement(ctx)
		return &InputSchemaStepIdentify{
			FlowRootObject:            flowRootObject,
			JSONPointer:               i.JSONPointer,
			Options:                   i.Options,
			ShouldBypassBotProtection: shouldBypassBotProtection,
			BotProtectionCfg:          deps.Config.BotProtection,
		}, nil
	case userIdentified && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentReauthFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	_, _, userIdentified := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseUser](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case len(flows.Nearest.Nodes) == 0 && authflow.GetIDToken(ctx) != "":
		identification := config.AuthenticationFlowIdentificationIDToken
		idx, err := i.checkIdentificationMethod(deps, step, identification)
		if err != nil {
			return nil, err
		}

		switch identification {
		case config.AuthenticationFlowIdentificationIDToken:
			return authflow.NewNodeSimple(&NodeIdentifyWithIDToken{
				JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
				Identification: identification,
			}), nil
		}
	case len(flows.Nearest.Nodes) == 0:
		var inputTakeIdentificationMethod inputTakeIdentificationMethod
		if authflow.AsInput(input, &inputTakeIdentificationMethod) {
			identification := inputTakeIdentificationMethod.GetIdentificationMethod()
			idx, err := i.checkIdentificationMethod(deps, step, identification)
			if err != nil {
				return nil, err
			}

			switch identification {
			case config.AuthenticationFlowIdentificationIDToken:
				return authflow.NewNodeSimple(&NodeIdentifyWithIDToken{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
				}), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	case userIdentified && !nestedStepsHandled:
		identification := i.identificationMethod(flows)
		return authflow.NewSubFlow(&IntentReauthFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, identification),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentReauthFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewIdentificationData(IdentificationData{
		Options: i.Options,
	}), nil
}

func (*IntentReauthFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowReauthFlowStep {
	step, ok := o.(*config.AuthenticationFlowReauthFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentReauthFlowStepIdentify) checkIdentificationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowReauthFlowStep, im config.AuthenticationFlowIdentification) (idx int, err error) {
	idx = -1

	for index, branch := range step.OneOf {
		branch := branch
		if im == branch.Identification {
			idx = index
		}
	}

	if idx >= 0 {
		return
	}

	err = authflow.ErrIncompatibleInput
	return
}

func (*IntentReauthFlowStepIdentify) identificationMethod(flows authflow.Flows) config.AuthenticationFlowIdentification {
	m, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneIdentificationMethod](flows)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneIdentificationMethod()

	return im
}

func (i *IntentReauthFlowStepIdentify) jsonPointer(step *config.AuthenticationFlowReauthFlowStep, im config.AuthenticationFlowIdentification) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (i *IntentReauthFlowStepIdentify) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	rootObject, err := flowRootObject(deps, i.FlowReference)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	return current, nil
}
