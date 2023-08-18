package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

func init() {
	workflow.RegisterIntent(&IntentSignupFlowStepUserProfile{})
}

type IntentSignupFlowStepUserProfile struct {
	SignupFlow  string        `json:"signup_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentSignupFlowStepUserProfile{}

func (i *IntentSignupFlowStepUserProfile) GetID() string {
	return i.StepID
}

func (i *IntentSignupFlowStepUserProfile) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ workflow.Intent = &IntentSignupFlowStepUserProfile{}
var _ workflow.Boundary = &IntentSignupFlowStepUserProfile{}

func (*IntentSignupFlowStepUserProfile) Kind() string {
	return "workflowconfig.IntentSignupFlowStepUserProfile"
}

func (i *IntentSignupFlowStepUserProfile) Boundary() string {
	return i.JSONPointer.String()
}

func (i *IntentSignupFlowStepUserProfile) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	if len(workflows.Nearest.Nodes) == 0 {
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
	return nil, workflow.ErrEOF
}

func (i *IntentSignupFlowStepUserProfile) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputFillUserProfile inputFillUserProfile
	if workflow.AsInput(input, &inputFillUserProfile) {
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

		stdAttrs, customAttrs, err := i.separate(deps, attributes)
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeDoUpdateUserProfile{
			UserID:             i.UserID,
			StandardAttributes: stdAttrs,
			CustomAttributes:   customAttrs,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentSignupFlowStepUserProfile) validate(step *config.WorkflowSignupFlowStep, attributes []attrs.T) (absent []string, err error) {
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
		return nil, InvalidUserProfile.NewWithInfo("invalid attributes", apierrors.Details{
			"allowed":    allAllowed,
			"required":   allRequired,
			"present":    allPresent,
			"absent":     allAbsent,
			"missing":    allMissing,
			"disallowed": allDisallowed,
		})
	}

	absent = allAbsent
	return
}

func (*IntentSignupFlowStepUserProfile) addAbsent(attributes []attrs.T, allAbsent []string) attrs.List {
	return attrs.List(attributes).AddAbsent(allAbsent)
}

func (*IntentSignupFlowStepUserProfile) separate(deps *workflow.Dependencies, attributes attrs.List) (stdAttrs attrs.List, customAttrs attrs.List, err error) {
	stdAttrs, customAttrs, unknownAttrs := attrs.List(attributes).Separate(deps.Config.UserProfile)
	if len(unknownAttrs) > 0 {
		err = InvalidUserProfile.NewWithInfo("unknown attributes", apierrors.Details{
			"unknown": unknownAttrs,
		})
	}
	return
}

func (*IntentSignupFlowStepUserProfile) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}
