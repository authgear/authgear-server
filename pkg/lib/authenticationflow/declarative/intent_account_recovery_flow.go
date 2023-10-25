package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterFlow(&IntentAccountRecoveryFlow{})
}

type IntentAccountRecoveryFlow struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentAccountRecoveryFlow{}
var _ authflow.EffectGetter = &IntentAccountRecoveryFlow{}

func (*IntentAccountRecoveryFlow) Kind() string {
	return "IntentAccountRecoveryFlow"
}

func (*IntentAccountRecoveryFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeAccountRecovery
}

func (i *IntentAccountRecoveryFlow) FlowInit(r authflow.FlowReference) {
	i.FlowReference = r
}

func (i *IntentAccountRecoveryFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentAccountRecoveryFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return flowRootObject(deps, i.FlowReference)
}

func (*IntentAccountRecoveryFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch {
	case len(flows.Nearest.Nodes) == 1:
		return nil, authflow.ErrEOF
	default:
		return nil, nil
	}
}

func (i *IntentAccountRecoveryFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return authflow.NewSubFlow(&IntentAccountRecoveryFlowSteps{
			JSONPointer: i.JSONPointer,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentAccountRecoveryFlow) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{}, nil
}
