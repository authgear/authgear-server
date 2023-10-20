package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentRequestAccountRecoveryFlowStepIdentify{})
}

type IntentRequestAccountRecoveryFlowStepIdentifyData struct {
	Options []AccountRecoveryIdentificationOption `json:"options"`
}

var _ authflow.Data = IntentRequestAccountRecoveryFlowStepIdentifyData{}

func (IntentRequestAccountRecoveryFlowStepIdentifyData) Data() {}

type IntentRequestAccountRecoveryFlowStepIdentify struct {
	JSONPointer jsonpointer.T                         `json:"json_pointer,omitempty"`
	StepName    string                                `json:"step_name,omitempty"`
	Options     []AccountRecoveryIdentificationOption `json:"options"`
}

var _ authflow.TargetStep = &IntentRequestAccountRecoveryFlowStepIdentify{}

func (i *IntentRequestAccountRecoveryFlowStepIdentify) GetName() string {
	return i.StepName
}

func (i *IntentRequestAccountRecoveryFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentRequestAccountRecoveryFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentRequestAccountRecoveryFlowStepIdentify{}

func NewIntentRequestAccountRecoveryFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentRequestAccountRecoveryFlowStepIdentify) (*IntentRequestAccountRecoveryFlowStepIdentify, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options := []AccountRecoveryIdentificationOption{}
	for _, b := range step.OneOf {
		switch b.Identification {
		case config.AuthenticationFlowRequestAccountRecoveryIdentificationEmail:
			fallthrough
		case config.AuthenticationFlowRequestAccountRecoveryIdentificationPhone:
			c := AccountRecoveryIdentificationOption{Identification: b.Identification}
			options = append(options, c)
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentRequestAccountRecoveryFlowStepIdentify) Kind() string {
	return "IntentRequestAccountRecoveryFlowStepIdentify"
}

func (i *IntentRequestAccountRecoveryFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		return &InputSchemaStepAccountRecoveryIdentify{
			JSONPointer: i.JSONPointer,
			Options:     i.Options,
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

func (i *IntentRequestAccountRecoveryFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeAccountRecoveryIdentificationMethod inputTakeAccountRecoveryIdentificationMethod
		if authflow.AsInput(input, &inputTakeAccountRecoveryIdentificationMethod) {
			identification := inputTakeAccountRecoveryIdentificationMethod.GetAccountRecoveryIdentificationMethod()
			idx, err := i.checkIdentificationMethod(deps, step, identification)
			if err != nil {
				return nil, err
			}
			branch := step.OneOf[idx]

			switch identification {
			case config.AuthenticationFlowRequestAccountRecoveryIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowRequestAccountRecoveryIdentificationPhone:
				return authflow.NewNodeSimple(&NodeUseAccountRecoveryIdentity{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
					OnFailure:      branch.OnFailure,
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
		return authflow.NewSubFlow(&IntentRequestAccountRecoveryFlowSteps{
			JSONPointer: i.jsonPointer(step, identification),
		}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentRequestAccountRecoveryFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return IntentRequestAccountRecoveryFlowStepIdentifyData{
		Options: i.Options,
	}, nil
}

func (*IntentRequestAccountRecoveryFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowRequestAccountRecoveryFlowStep {
	step, ok := o.(*config.AuthenticationFlowRequestAccountRecoveryFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}

func (*IntentRequestAccountRecoveryFlowStepIdentify) checkIdentificationMethod(
	deps *authflow.Dependencies,
	step *config.AuthenticationFlowRequestAccountRecoveryFlowStep,
	im config.AuthenticationFlowRequestAccountRecoveryIdentification,
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

func (*IntentRequestAccountRecoveryFlowStepIdentify) identificationMethod(w *authflow.Flow) config.AuthenticationFlowRequestAccountRecoveryIdentification {
	m, ok := authflow.FindMilestone[MilestoneAccountRecoveryIdentificationMethod](w)
	if !ok {
		panic(fmt.Errorf("identification method not yet selected"))
	}

	im := m.MilestoneAccountRecoveryIdentificationMethod()

	return im
}

func (i *IntentRequestAccountRecoveryFlowStepIdentify) jsonPointer(
	step *config.AuthenticationFlowRequestAccountRecoveryFlowStep,
	im config.AuthenticationFlowRequestAccountRecoveryIdentification,
) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Identification == im {
			return authflow.JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("selected identification method is not allowed"))
}
