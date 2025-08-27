package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterFlow(&IntentSignupLoginFlow{})
}

type IntentSignupLoginFlow struct {
	FlowReference authflow.FlowReference `json:"flow_reference,omitempty"`
	JSONPointer   jsonpointer.T          `json:"json_pointer,omitempty"`
}

var _ authflow.PublicFlow = &IntentSignupLoginFlow{}

func (*IntentSignupLoginFlow) Kind() string {
	return "IntentSignupLoginFlow"
}

func (*IntentSignupLoginFlow) FlowType() authflow.FlowType {
	return authflow.FlowTypeSignupLogin
}

func (i *IntentSignupLoginFlow) FlowInit(r authflow.FlowReference, startFrom jsonpointer.T) {
	i.FlowReference = r
}

func (i *IntentSignupLoginFlow) FlowFlowReference() authflow.FlowReference {
	return i.FlowReference
}

func (i *IntentSignupLoginFlow) FlowRootObject(deps *authflow.Dependencies) (config.AuthenticationFlowObject, error) {
	return GetFlowRootObject(deps.Config, i.FlowReference)
}

func (i *IntentSignupLoginFlow) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return nil, nil
	case len(flows.Nearest.Nodes) == 1:
		return nil, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentSignupLoginFlow) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	switch {
	case len(flows.Nearest.Nodes) == 0:
		return NewNodePreInitialize(ctx, deps, flows)
	case len(flows.Nearest.Nodes) == 1:
		return authflow.NewSubFlow(&IntentSignupLoginFlowSteps{
			FlowReference: i.FlowReference,
			JSONPointer:   i.JSONPointer,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
