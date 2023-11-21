package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentSignupLoginFlowStepIdentify{})
}

type intentSignupLoginFlowStepIdentifyData struct {
	Options []IdentificationOption `json:"options"`
}

var _ authflow.Data = intentSignupLoginFlowStepIdentifyData{}

func (intentSignupLoginFlowStepIdentifyData) Data() {}

type IntentSignupLoginFlowStepIdentify struct {
	JSONPointer jsonpointer.T          `json:"json_pointer,omitempty"`
	StepName    string                 `json:"step_name,omitempty"`
	Options     []IdentificationOption `json:"options"`
}

var _ authflow.Intent = &IntentSignupLoginFlowStepIdentify{}
var _ authflow.DataOutputer = &IntentSignupLoginFlowStepIdentify{}

func NewIntentSignupLoginFlowStepIdentify(ctx context.Context, deps *authflow.Dependencies, i *IntentSignupLoginFlowStepIdentify) (*IntentSignupLoginFlowStepIdentify, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	options := []IdentificationOption{}
	for _, b := range step.OneOf {
		switch b.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			fallthrough
		case config.AuthenticationFlowIdentificationPhone:
			fallthrough
		case config.AuthenticationFlowIdentificationUsername:
			c := NewIdentificationOptionLoginID(b.Identification)
			options = append(options, c)
		case config.AuthenticationFlowIdentificationOAuth:
			oauthOptions := NewIdentificationOptionsOAuth(
				deps.Config.Identity.OAuth,
				deps.FeatureConfig.Identity.OAuth.Providers,
			)
			options = append(options, oauthOptions...)
		case config.AuthenticationFlowIdentificationPasskey:
			// Passkey is for login only.
			requestOptions, err := deps.PasskeyRequestOptionsService.MakeModalRequestOptions()
			if err != nil {
				return nil, err
			}
			c := NewIdentificationOptionPasskey(requestOptions)
			options = append(options, c)
		}
	}

	i.Options = options
	return i, nil
}

func (*IntentSignupLoginFlowStepIdentify) Kind() string {
	return "IntentSignupLoginFlowStepIdentify"
}

func (i *IntentSignupLoginFlowStepIdentify) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Let the input to select which identification method to use.
	if len(flows.Nearest.Nodes) == 0 {
		return &InputSchemaStepIdentify{
			JSONPointer: i.JSONPointer,
			Options:     i.Options,
		}, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentSignupLoginFlowStepIdentify) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	if len(flows.Nearest.Nodes) == 0 {
		var inputTakeIdentificationMethod inputTakeIdentificationMethod
		if authflow.AsInput(input, &inputTakeIdentificationMethod) {
			identification := inputTakeIdentificationMethod.GetIdentificationMethod()
			idx, err := i.checkIdentificationMethod(deps, step, identification)
			if err != nil {
				return nil, err
			}

			syntheticInput := &InputStepIdentify{
				Identification: identification,
			}

			switch identification {
			case config.AuthenticationFlowIdentificationEmail:
				fallthrough
			case config.AuthenticationFlowIdentificationPhone:
				fallthrough
			case config.AuthenticationFlowIdentificationUsername:
				return authflow.NewNodeSimple(&NodeLookupIdentityLoginID{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
					SyntheticInput: syntheticInput,
				}), nil
			case config.AuthenticationFlowIdentificationOAuth:
				return authflow.NewSubFlow(&IntentLookupIdentityOAuth{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
					SyntheticInput: syntheticInput,
				}), nil
			case config.AuthenticationFlowIdentificationPasskey:
				return authflow.NewNodeSimple(&NodeLookupIdentityPasskey{
					JSONPointer:    authflow.JSONPointerForOneOf(i.JSONPointer, idx),
					Identification: identification,
					SyntheticInput: syntheticInput,
				}), nil
			}
		}
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentSignupLoginFlowStepIdentify) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return intentSignupLoginFlowStepIdentifyData{
		Options: i.Options,
	}, nil
}

func (i *IntentSignupLoginFlowStepIdentify) checkIdentificationMethod(deps *authflow.Dependencies, step *config.AuthenticationFlowSignupLoginFlowStep, im config.AuthenticationFlowIdentification) (idx int, err error) {
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

func (i *IntentSignupLoginFlowStepIdentify) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}
