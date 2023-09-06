package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

func init() {
	authflow.RegisterIntent(&IntentSignupFlowStepUserProfile{})
}

type IntentSignupFlowStepUserProfile struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ FlowStep = &IntentSignupFlowStepUserProfile{}

func (i *IntentSignupFlowStepUserProfile) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepUserProfile) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ authflow.Intent = &IntentSignupFlowStepUserProfile{}

func (*IntentSignupFlowStepUserProfile) Kind() string {
	return "IntentSignupFlowStepUserProfile"
}

func (i *IntentSignupFlowStepUserProfile) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if len(flows.Nearest.Nodes) == 0 {
		current, err := signupFlowCurrent(deps, i.SignupFlow, i.JSONPointer)
		if err != nil {
			return nil, err
		}

		step := i.step(current)
		if err != nil {
			return nil, err
		}
		return &InputSchemaFillUserProfile{
			Attributes:       step.UserProfile,
			CustomAttributes: deps.Config.UserProfile.CustomAttributes.Attributes,
		}, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentSignupFlowStepUserProfile) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputFillUserProfile inputFillUserProfile
	if authflow.AsInput(input, &inputFillUserProfile) {
		current, err := signupFlowCurrent(deps, i.SignupFlow, i.JSONPointer)
		if err != nil {
			return nil, err
		}

		step := i.step(current)
		if err != nil {
			return nil, err
		}

		attributes := inputFillUserProfile.GetAttributes()
		allAbsent, err := i.validate(step, attributes)
		if err != nil {
			return nil, err
		}

		attributes = i.addAbsent(attributes, allAbsent)

		stdAttrs, customAttrs := i.separate(deps, attributes)

		return authflow.NewNodeSimple(&NodeDoUpdateUserProfile{
			UserID:             i.UserID,
			StandardAttributes: stdAttrs,
			CustomAttributes:   customAttrs,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (*IntentSignupFlowStepUserProfile) validate(step *config.AuthenticationFlowSignupFlowStep, attributes []attrs.T) (absent []string, err error) {
	allAllowed := []string{}
	allRequired := []string{}
	for _, spec := range step.UserProfile {
		spec := spec
		allAllowed = append(allAllowed, spec.Pointer)
		if spec.Required {
			allRequired = append(allRequired, spec.Pointer)
		}
	}

	allPresent := []string{}
	for _, attr := range attributes {
		attr := attr
		pointer := attr.Pointer
		allPresent = append(allPresent, pointer)
	}

	allMissing := slice.ExceptStrings(allRequired, allPresent)
	allDisallowed := slice.ExceptStrings(allPresent, allAllowed)
	allAbsent := slice.ExceptStrings(allAllowed, allPresent)

	if len(allMissing) > 0 || len(allDisallowed) > 0 {
		panic(fmt.Errorf("the input schema should have ensured there are missing or disallowed attributes"))
	}

	absent = allAbsent
	return
}

func (*IntentSignupFlowStepUserProfile) addAbsent(attributes []attrs.T, allAbsent []string) attrs.List {
	return attrs.List(attributes).AddAbsent(allAbsent)
}

func (*IntentSignupFlowStepUserProfile) separate(deps *authflow.Dependencies, attributes attrs.List) (stdAttrs attrs.List, customAttrs attrs.List) {
	stdAttrs, customAttrs, unknownAttrs := attrs.List(attributes).Separate(deps.Config.UserProfile)
	if len(unknownAttrs) > 0 {
		panic(fmt.Errorf("the input schema should have ensured there are no unknown attributes"))
	}
	return
}

func (*IntentSignupFlowStepUserProfile) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}
