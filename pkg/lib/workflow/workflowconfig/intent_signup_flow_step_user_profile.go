package workflowconfig

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentSignupFlowStepUserProfile{})
}

var IntentSignupFlowStepUserProfileSchema = validation.NewSimpleSchema(`{}`)

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

func (*IntentSignupFlowStepUserProfile) Kind() string {
	return "workflowconfig.IntentSignupFlowStepUserProfile"
}

func (*IntentSignupFlowStepUserProfile) JSONSchema() *validation.SimpleSchema {
	return IntentSignupFlowStepUserProfileSchema
}

func (*IntentSignupFlowStepUserProfile) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputFillUserProfile{}}, nil
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
		err = i.validate(step, attributes)
		if err != nil {
			return nil, err
		}

		// FIXME(workflow): separate attributes into standard attributes and custom attributes.
		// FIXME(workflow): update standard attributes
		// FIXME(workflow): update custom attributes.
		return workflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentSignupFlowStepUserProfile) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentSignupFlowStepUserProfile) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentSignupFlowStepUserProfile) validate(step *config.WorkflowSignupFlowStep, attributes []InputFillUserProfileAttribute) error {
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
		pointer := attr.Pointer.String()
		allPresent = append(allPresent, pointer)
	}

	allMissing := slice.ExceptStrings(allRequired, allPresent)
	allDisallowed := slice.ExceptStrings(allPresent, allAllowed)

	if len(allMissing) > 0 || len(allDisallowed) > 0 {
		return InvalidUserProfile.NewWithInfo("invalid attributes", apierrors.Details{
			"allowed":    allAllowed,
			"required":   allRequired,
			"actual":     allPresent,
			"missing":    allMissing,
			"disallowed": allDisallowed,
		})
	}

	return nil
}

func (*IntentSignupFlowStepUserProfile) step(o config.WorkflowObject) *config.WorkflowSignupFlowStep {
	step, ok := o.(*config.WorkflowSignupFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}
