package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterFlow(&IntentRequestAccountRecoveryFlow{})
}

type IntentRequestAccountRecoveryFlow struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentRequestAccountRecoveryFlow{}
var _ authflow.EffectGetter = &IntentRequestAccountRecoveryFlow{}

func (*IntentRequestAccountRecoveryFlow) Kind() string {
	return "IntentRequestAccountRecoveryFlow"
}

func (*IntentRequestAccountRecoveryFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeRequestAccountRecovery
}

func (i *IntentRequestAccountRecoveryFlow) FlowInit(r authflow.FlowReference) {
	i.FlowReference = r
}

func (i *IntentRequestAccountRecoveryFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentRequestAccountRecoveryFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return flowRootObject(deps, i.FlowReference)
}

func (*IntentRequestAccountRecoveryFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch {
	case len(flows.Nearest.Nodes) == 1:
		return nil, authflow.ErrEOF
	default:
		return nil, nil
	}
}

func (i *IntentRequestAccountRecoveryFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return authflow.NewSubFlow(&IntentRequestAccountRecoveryFlowSteps{
			JSONPointer: i.JSONPointer,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentRequestAccountRecoveryFlow) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]authflow.Effect, error) {
	return []authflow.Effect{}, nil
}
