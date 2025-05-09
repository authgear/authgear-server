package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentIdentifyWithIDToken{})
}

type IntentIdentifyWithIDToken struct {
	JSONPointer      jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification   config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	PrefilledIDToken string                                  `json:"id_token,omitempty"`
}

var _ authflow.Intent = &IntentIdentifyWithIDToken{}
var _ authflow.Milestone = &IntentIdentifyWithIDToken{}
var _ MilestoneIdentificationMethod = &IntentIdentifyWithIDToken{}
var _ MilestoneFlowUseIdentity = &IntentIdentifyWithIDToken{}
var _ authflow.InputReactor = &IntentIdentifyWithIDToken{}

func (*IntentIdentifyWithIDToken) Kind() string {
	return "IntentIdentifyWithIDToken"
}

func (*IntentIdentifyWithIDToken) Milestone() {}
func (n *IntentIdentifyWithIDToken) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}
func (*IntentIdentifyWithIDToken) MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool) {
	// This intent does not contain any node that implements MilestoneDoUseIdentity because the ID token is not associated with an identity.
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
}

func (n *IntentIdentifyWithIDToken) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, userIdentified := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseUser](flows)
	if userIdentified {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	switch {
	case n.PrefilledIDToken != "":
		// Special case: if id_token is available, use it automatically.
		return nil, nil
	default:
		return &InputSchemaTakeIDToken{
			FlowRootObject: flowRootObject,
			JSONPointer:    n.JSONPointer,
		}, nil
	}
}

func (n *IntentIdentifyWithIDToken) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	proceed := func(idToken string) (*authflow.Node, error) {
		n, err := NewNodeDoUseIDToken(ctx, deps, flows, &NodeDoUseIDToken{
			IDToken: idToken,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), nil
	}

	switch {
	case n.PrefilledIDToken != "":
		return proceed(n.PrefilledIDToken)
	default:
		var inputTakeIDToken inputTakeIDToken
		if authflow.AsInput(input, &inputTakeIDToken) {
			idToken := inputTakeIDToken.GetIDToken()
			return proceed(idToken)
		}

		return nil, authflow.ErrIncompatibleInput
	}
}
