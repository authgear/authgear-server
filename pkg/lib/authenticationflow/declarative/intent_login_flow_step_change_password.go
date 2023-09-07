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
	GetPasswordAuthenticator(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (*authenticator.Info, bool)
}

func init() {
	authflow.RegisterIntent(&IntentLoginFlowStepChangePassword{})
}

type IntentLoginFlowStepChangePassword struct {
	StepID      string        `json:"step_id,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ FlowStep = &IntentLoginFlowStepChangePassword{}

func (i *IntentLoginFlowStepChangePassword) GetID() string {
	return i.StepID
}

func (i *IntentLoginFlowStepChangePassword) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
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
	current, err := authflow.FlowObject(authflow.GetFlowRootObject(ctx), i.JSONPointer)
	if err != nil {
		return nil, err
	}

	step := i.step(current)
	targetStepID := step.TargetStep

	targetStepFlow, err := FindTargetStep(flows.Root, targetStepID)
	if err != nil {
		return nil, err
	}

	target, ok := targetStepFlow.Intent.(IntentLoginFlowStepChangePasswordTarget)
	if !ok {
		return nil, InvalidTargetStep.NewWithInfo("invalid target_step", apierrors.Details{
			"target_step": targetStepID,
		})
	}

	info, ok := target.GetPasswordAuthenticator(ctx, deps, flows.Replace(targetStepFlow))
	if !ok {
		// No need to change. End this flow.
		return authflow.NewNodeSimple(&NodeSentinel{}), nil
	}

	return authflow.NewNodeSimple(&NodeLoginFlowChangePassword{
		Authenticator: info,
	}), nil
}

func (*IntentLoginFlowStepChangePassword) step(o config.AuthenticationFlowObject) *config.AuthenticationFlowLoginFlowStep {
	step, ok := o.(*config.AuthenticationFlowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return step
}
