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
	authflow.RegisterIntent(&IntentSignupFlowStepFillInUserProfile{})
}

type IntentSignupFlowStepFillInUserProfile struct {
	JSONPointer            jsonpointer.T `json:"json_pointer,omitempty"`
	StepName               string        `json:"step_name,omitempty"`
	UserID                 string        `json:"user_id,omitempty"`
	IsUpdatingExistingUser bool          `json:"skip_update,omitempty"`
}

var _ authflow.Intent = &IntentSignupFlowStepFillInUserProfile{}
var _ authflow.Milestone = &IntentSignupFlowStepFillInUserProfile{}
var _ MilestoneSwitchToExistingUser = &IntentSignupFlowStepFillInUserProfile{}

func (*IntentSignupFlowStepFillInUserProfile) Kind() string {
	return "IntentSignupFlowStepFillInUserProfile"
}

func (*IntentSignupFlowStepFillInUserProfile) Milestone() {}
func (i *IntentSignupFlowStepFillInUserProfile) MilestoneSwitchToExistingUser(deps *authflow.Dependencies, flow *authflow.Flow, newUserID string) error {
	i.UserID = newUserID
	i.IsUpdatingExistingUser = true
	return nil
}

func (i *IntentSignupFlowStepFillInUserProfile) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if !i.IsUpdatingExistingUser && len(flows.Nearest.Nodes) == 0 {
		flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}

		current, err := authflow.FlowObject(flowRootObject, i.JSONPointer)
		if err != nil {
			return nil, err
		}

		step := i.step(current)
		return &InputSchemaFillInUserProfile{
			JSONPointer:      i.JSONPointer,
			FlowRootObject:   flowRootObject,
			Attributes:       step.UserProfile,
			CustomAttributes: deps.Config.UserProfile.CustomAttributes.Attributes,
		}, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentSignupFlowStepFillInUserProfile) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputFillInUserProfile inputFillInUserProfile
	if !i.IsUpdatingExistingUser && authflow.AsInput(input, &inputFillInUserProfile) {
		rootObject, err := findFlowRootObjectInFlow(deps, flows)
		if err != nil {
			return nil, err
		}
		current, err := authflow.FlowObject(rootObject, i.JSONPointer)
		if err != nil {
			return nil, err
		}

		step := i.step(current)

		attributes := inputFillInUserProfile.GetAttributes()
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

func (*IntentSignupFlowStepFillInUserProfile) validate(step *config.AuthenticationFlowSignupFlowStep, attributes []attrs.T) (absent []string, err error) {
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

func (*IntentSignupFlowStepFillInUserProfile) addAbsent(attributes []attrs.T, allAbsent []string) attrs.List {
	return attrs.List(attributes).AddAbsent(allAbsent)
}

func (*IntentSignupFlowStepFillInUserProfile) separate(deps *authflow.Dependencies, attributes attrs.List) (stdAttrs attrs.List, customAttrs attrs.List) {
	stdAttrs, customAttrs, unknownAttrs := attrs.List(attributes).Separate(deps.Config.UserProfile)
	if len(unknownAttrs) > 0 {
		panic(fmt.Errorf("the input schema should have ensured there are no unknown attributes"))
	}
	return
}

func (*IntentSignupFlowStepFillInUserProfile) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupFlowStep {
	step, ok := o.(*config.AuthenticationFlowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}
