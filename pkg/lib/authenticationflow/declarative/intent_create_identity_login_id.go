package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

func init() {
	authflow.RegisterIntent(&IntentCreateIdentityLoginID{})
}

type IntentCreateIdentityLoginID struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.Intent = &IntentCreateIdentityLoginID{}
var _ authflow.Milestone = &IntentCreateIdentityLoginID{}
var _ MilestoneIdentificationMethod = &IntentCreateIdentityLoginID{}
var _ MilestoneFlowCreateIdentity = &IntentCreateIdentityLoginID{}
var _ authflow.InputReactor = &IntentCreateIdentityLoginID{}

func (*IntentCreateIdentityLoginID) Milestone() {}

func (*IntentCreateIdentityLoginID) Kind() string {
	return "IntentCreateIdentityLoginID"
}

func (n *IntentCreateIdentityLoginID) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *IntentCreateIdentityLoginID) MilestoneFlowCreateIdentity(flows authflow.Flows) (MilestoneDoCreateIdentity, authflow.Flows, bool) {
	// Find IntentCheckConflictAndCreateIdenity
	m, mFlows, ok := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	if !ok {
		return nil, mFlows, false
	}

	// Delegate to IntentCheckConflictAndCreateIdenity
	return m.MilestoneFlowCreateIdentity(mFlows)
}

func (n *IntentCreateIdentityLoginID) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, identified := authflow.FindMilestoneInCurrentFlow[MilestoneFlowCreateIdentity](flows)
	if identified {
		return nil, authflow.ErrEOF
	}

	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer, n)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeLoginID{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}

func (n *IntentCreateIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}

		loginID := inputTakeLoginID.GetLoginID()
		spec := makeLoginIDSpec(n.Identification, stringutil.NewUserInputString(loginID))

		return authflow.NewSubFlow(&IntentCheckConflictAndCreateIdenity{
			JSONPointer: n.JSONPointer,
			UserID:      n.UserID,
			Request:     NewCreateLoginIDIdentityRequest(spec),
		}), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
