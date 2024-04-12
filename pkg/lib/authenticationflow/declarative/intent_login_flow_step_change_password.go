package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type IntentLoginFlowStepChangePasswordTarget interface {
	GetChangeRequiredPasswordAuthenticator(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (*authenticator.Info, PasswordChangeReason)
}

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepChangePassword{})
}

type IntentLoginFlowStepChangePassword struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	StepName      string                 `json:"step_name,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
}

var _ authflow.Intent = &IntentLoginFlowStepChangePassword{}

func (*IntentLoginFlowStepChangePassword) Kind() string {
	return "IntentLoginFlowStepChangePassword"
}

func (*IntentLoginFlowStepChangePassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// Look up the password authenticator to change.
	if len(flows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentLoginFlowStepChangePassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, _ authflow.Input) (*authflow.Node, error) {
	current, err := i.currentFlowObject(deps)
	if err != nil {
		return nil, err
	}

	step := i.step(current)
	targetStepName := step.TargetStep

	targetStepFlow, err := authflow.FindTargetStep(flows.Root, targetStepName)
	if err != nil {
		return nil, err
	}

	target, ok := targetStepFlow.Intent.(IntentLoginFlowStepChangePasswordTarget)
	if !ok {
		return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
			"target_step": targetStepName,
		})
	}

	info, changeReason := target.GetChangeRequiredPasswordAuthenticator(ctx, deps, flows.Replace(targetStepFlow))
	if info == nil {
		// No need to change. End this flow.
		return authflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	switch changeReason {
	case PasswordChangeReasonExpiry:
		// Always force change password if it is expired.
	case PasswordChangeReasonPolicy:
		fallthrough
	default:
		if !*deps.Config.Authenticator.Password.ForceChange {
			return authflow.NewNodeSimple(&NodeSentinel{}), nil
		}
	}

	return authflow.NewNodeSimple(&NodeLoginFlowChangePassword{
		JSONPointer:   i.JSONPointer,
		Authenticator: info,
		Reason:        &changeReason,
	}), nil
}

func (i *IntentLoginFlowStepChangePassword) currentFlowObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
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

func (*IntentLoginFlowStepChangePassword) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}
