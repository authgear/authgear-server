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
	StartFrom     jsonpointer.T          `json:"start_from,omitempty"`
}

var _ authflow.PublicFlow = &IntentAccountRecoveryFlow{}

func (*IntentAccountRecoveryFlow) Kind() string {
	return "IntentAccountRecoveryFlow"
}

func (*IntentAccountRecoveryFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeAccountRecovery
}

func (i *IntentAccountRecoveryFlow) FlowInit(r authflow.FlowReference, startFrom jsonpointer.T) {
	i.FlowReference = r
	i.StartFrom = startFrom
}

func (i *IntentAccountRecoveryFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentAccountRecoveryFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return GetFlowRootObject(deps.Config, i.FlowReference)
}

func (*IntentAccountRecoveryFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch {
	case len(flows.Nearest.Nodes) == 1:
		return nil, authflow.ErrEOF
	default:
		return nil, nil
	}
}

func (i *IntentAccountRecoveryFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return authflow.NewSubFlow(&IntentAccountRecoveryFlowSteps{
			JSONPointer:   i.JSONPointer,
			FlowReference: i.FlowReference,
			StartFrom:     i.StartFrom,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
