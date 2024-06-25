package declarative

import (
	"context"
	"fmt"
	"strconv"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentAccountRecoveryFlowStepIdentify{})
}

// IntentAccountRecoveryFlowStepIdentify
//
//   NodeSentinel
//
//   NodeUseAccountRecoveryIdentity (MilestoneDoUseAccountRecoveryIdentificationMethod)
//   NodeDoUseAccountRecoveryIdentity (MilestoneDoUseAccountRecoveryIdentity)

type IntentAccountRecoveryFlowStepIdentifyData struct {
	TypedData
	Options []AccountRecoveryIdentificationOption `json:"options"`
}

func NewIntentAccountRecoveryFlowStepIdentifyData(d IntentAccountRecoveryFlowStepIdentifyData) IntentAccountRecoveryFlowStepIdentifyData {
	d.Type = DataTypeAccountRecoveryIdentificationData
	return d
}

var _ authflow.Data = IntentAccountRecoveryFlowStepIdentifyData{}

func (IntentAccountRecoveryFlowStepIdentifyData) Data() {}

type IntentAccountRecoveryFlowStepIdentify struct {
	FlowReference authflow.FlowReference                `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T                         `json:"json_pointer,omitempty"`
	StepName      string                                `json:"step_name,omitempty"`
	Options       []AccountRecoveryIdentificationOption `json:"options"`
	StartFrom     jsonpointer.T                         `json:"start_from,omitempty"`
}

var _ authflow.TargetStep = &IntentAccountRecoveryFlowStepIdentify{}

func (i *IntentAccountRecoveryFlowStepIdentify) GetName() string {
	return i.StepName
}

func (i *IntentAccountRecoveryFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentAccountRecoveryFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentAccountRecoveryFlowStepIdentify{}

func NewIntentAccountRecoveryFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentAccountRecoveryFlowStepIdentify) (*IntentAccountRecoveryFlowStepIdentify, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options := []AccountRecoveryIdentificationOption{}
	for _, b := range step.OneOf {
		switch b.Identification {
		case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
			fallthrough
		case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
			c := AccountRecoveryIdentificationOption{Identification: b.Identification}
			options = append(options, c)
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentAccountRecoveryFlowStepIdentify) Kind() string {
	return "IntentAccountRecoveryFlowStepIdentify"
}

func (i *IntentAccountRecoveryFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, isSelectedIdenRestored := i.restoredIdentificationJsonPointer()
	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		if isSelectedIdenRestored {
			// When restoring the intent, no input is needed
			return nil, nil
		}
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaStepAccountRecoveryIdentify{
			FlowRootObject: flowRootObject,
			JSONPointer:    i.JSONPointer,
			Options:        i.Options,
		}, nil
	}

	_, _, identityUsed := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseAccountRecoveryIdentificationMethod](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)

	switch {
	case (identityUsed || isSelectedIdenRestored) && !nestedStepsHandled:
		// Handle nested steps.
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentAccountRecoveryFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		if i.isRestored() {
			// When restoring the intent, skip the first node
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}
		var inputTakeAccountRecoveryIdentificationMethod inputTakeAccountRecoveryIdentificationMethod
		if authflow.AsInput(input, &inputTakeAccountRecoveryIdentificationMethod) {
			identification := inputTakeAccountRecoveryIdentificationMethod.GetAccountRecoveryIdentificationMethod()
			idx, err := i.checkIdentificationMethod(deps, step, identification)
			if err != nil {
				return nil, err
			}
			branch := step.OneOf[idx]

			switch identification {
			case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
				return authflow.NewNodeSimple(&NodeUseAccountRecoveryIdentity{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
					OnFailure:      branch.OnFailure,
				}), nil
			}
		}
		return nil, authflow.ErrIncompatibleInput
	}

	_, _, identityUsed := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseAccountRecoveryIdentificationMethod](flows)
	_, _, nestedStepsHandled := authflow.FindMilestoneInCurrentFlow[MilestoneNestedSteps](flows)
	restoredIdenJsonPointer, isSelectedIdenRestored := i.restoredIdentificationJsonPointer()

	switch {
	case isSelectedIdenRestored && !nestedStepsHandled:
		return authflow.NewSubFlow(&IntentAccountRecoveryFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   restoredIdenJsonPointer,
			StartFrom:     i.StartFrom,
		}), nil
	case identityUsed && !nestedStepsHandled:
		identification := i.identificationMethod(flows)
		return authflow.NewSubFlow(&IntentAccountRecoveryFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.jsonPointer(step, identification),
			StartFrom:     i.StartFrom,
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentAccountRecoveryFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewIntentAccountRecoveryFlowStepIdentifyData(IntentAccountRecoveryFlowStepIdentifyData{
		Options: i.Options,
	}), nil
}

func (*IntentAccountRecoveryFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowAccountRecoveryFlowStep {
	step, ok := o.(*config.AuthenticationFlowAccountRecoveryFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (i *IntentAccountRecoveryFlowStepIdentify) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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

func (*IntentAccountRecoveryFlowStepIdentify) checkIdentificationMethod(
	deps *authflow.Dependencies,
	step *config.AuthenticationFlowAccountRecoveryFlowStep,
	im config.AuthenticationFlowAccountRecoveryIdentification,
) (idx int, err error) {
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

func (*IntentAccountRecoveryFlowStepIdentify) identificationMethod(flows authflow.Flows) config.AuthenticationFlowAccountRecoveryIdentification {
	m, _, ok := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseAccountRecoveryIdentificationMethod](flows)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneDoUseAccountRecoveryIdentificationMethod()

	return im
}

func (i *IntentAccountRecoveryFlowStepIdentify) jsonPointer(
	step *config.AuthenticationFlowAccountRecoveryFlowStep,
	im config.AuthenticationFlowAccountRecoveryIdentification,
) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}

func (i *IntentAccountRecoveryFlowStepIdentify) isRestored() bool {
	return isNodeRestored(i.JSONPointer, i.StartFrom)
}

func (i *IntentAccountRecoveryFlowStepIdentify) restoredIdentificationJsonPointer() (jsonpointer.T, bool) {
	if !i.isRestored() {
		return nil, false
	}

	startFrom := authflow.JSONPointerSubtract(i.StartFrom, i.JSONPointer)
	if len(startFrom) < 2 || startFrom[0] != authflow.JsonPointerTokenOneOf {
		return nil, false
	}
	idenIdx, err := strconv.Atoi(startFrom[1])
	if err != nil {
		return nil, false
	}
	return authflow.JSONPointerForOneOf(i.JSONPointer, idenIdx), true
}
