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
	authflow.RegisterIntent(&IntentUseIdentityLoginID{})
}

type IntentUseIdentityLoginID struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.Intent = &IntentUseIdentityLoginID{}
var _ authflow.Milestone = &IntentUseIdentityLoginID{}
var _ MilestoneIdentificationMethod = &IntentUseIdentityLoginID{}
var _ MilestoneFlowUseIdentity = &IntentUseIdentityLoginID{}
var _ authflow.InputReactor = &IntentUseIdentityLoginID{}

func (*IntentUseIdentityLoginID) Kind() string {
	return "IntentUseIdentityLoginID"
}

func (*IntentUseIdentityLoginID) Milestone() {}
func (n *IntentUseIdentityLoginID) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}
func (*IntentUseIdentityLoginID) MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
}

func (n *IntentUseIdentityLoginID) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, identified := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
	if identified {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeLoginID{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
	}, nil
}

func (n *IntentUseIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		var bpSpecialErr error
		bpRequired, err := IsNodeBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
		if err != nil {
			return nil, err
		}
		if bpRequired {
			var inputTakeBotProtection inputTakeBotProtection
			if !authflow.AsInput(input, &inputTakeBotProtection) {
				return nil, authflow.ErrIncompatibleInput
			}

			token := inputTakeBotProtection.GetBotProtectionProviderResponse()
			bpSpecialErr, err = HandleBotProtection(ctx, deps, token)
			if err != nil {
				return nil, err
			}
		}
		loginID := inputTakeLoginID.GetLoginID()
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Value: loginID,
			},
		}

		exactMatch, err := findExactOneIdentityInfo(deps, spec)
		if err != nil {
			return nil, err
		}

		n, err := NewNodeDoUseIdentity(ctx, flows, &NodeDoUseIdentity{
			Identity: exactMatch,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
