package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentUseAccountRecoveryIdentity{})
}

type IntentUseAccountRecoveryIdentity struct {
	JSONPointer    jsonpointer.T                                                   `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowAccountRecoveryIdentification          `json:"identification,omitempty"`
	OnFailure      config.AuthenticationFlowAccountRecoveryIdentificationOnFailure `json:"on_failure,omitempty"`
}

var _ authflow.Intent = &IntentUseAccountRecoveryIdentity{}
var _ authflow.Milestone = &IntentUseAccountRecoveryIdentity{}
var _ MilestoneDoUseAccountRecoveryIdentificationMethod = &IntentUseAccountRecoveryIdentity{}
var _ authflow.InputReactor = &IntentUseAccountRecoveryIdentity{}

func (*IntentUseAccountRecoveryIdentity) Kind() string {
	return "IntentUseAccountRecoveryIdentity"
}

func (*IntentUseAccountRecoveryIdentity) Milestone() {}
func (n *IntentUseAccountRecoveryIdentity) MilestoneDoUseAccountRecoveryIdentificationMethod() config.AuthenticationFlowAccountRecoveryIdentification {
	return n.Identification
}

func (n *IntentUseAccountRecoveryIdentity) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakeLoginID{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *IntentUseAccountRecoveryIdentity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		loginID := inputTakeLoginID.GetLoginID()
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Value: loginID,
			},
		}

		var exactMatch *identity.Info = nil
		switch n.OnFailure {
		case config.AuthenticationFlowAccountRecoveryIdentificationOnFailureIgnore:
			var err error
			exactMatch, _, err = deps.Identities.SearchBySpec(spec)
			if err != nil {
				return nil, err
			}
		case config.AuthenticationFlowAccountRecoveryIdentificationOnFailureError:
			var err error
			exactMatch, err = findExactOneIdentityInfo(deps, spec)
			if err != nil {
				return nil, err
			}
		}

		nextNode := &NodeDoUseAccountRecoveryIdentity{
			Identification: n.Identification,
			Spec:           spec,
			MaybeIdentity:  exactMatch,
		}

		return authflow.NewNodeSimple(nextNode), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
