package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterIntent(&IntentUseIdentitySelectAccount{})
}

// IntentUseIdentitySelectAccount
//   NodeDoUseIdentitySelectAccount (MilestoneDoUseUser, not MilestoneDoUseIdentity!)

type IntentUseIdentitySelectAccount struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	ExpectedUserID string                                 `json:"expected_user_id,omitempty"`
}

var _ authflow.Intent = &IntentUseIdentitySelectAccount{}
var _ authflow.Milestone = &IntentUseIdentitySelectAccount{}
var _ MilestoneIdentificationMethod = &IntentUseIdentitySelectAccount{}
var _ MilestoneFlowUseIdentity = &IntentUseIdentitySelectAccount{}
var _ authflow.InputReactor = &IntentUseIdentitySelectAccount{}

func (*IntentUseIdentitySelectAccount) Kind() string {
	return "IntentUseIdentitySelectAccount"
}

func (*IntentUseIdentitySelectAccount) Milestone() {}

func (n *IntentUseIdentitySelectAccount) MilestoneIdentificationMethod() model.AuthenticationFlowIdentification {
	return n.Identification
}

func (*IntentUseIdentitySelectAccount) MilestoneFlowUseIdentity(flows authflow.Flows) (MilestoneDoUseIdentity, authflow.Flows, bool) {
	// This intent does not contain any node that implements MilestoneDoUseIdentity
	// because select_account reuses the session as-is; it is not associated
	// with any specific identity.
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoUseIdentity](flows)
}

func (n *IntentUseIdentitySelectAccount) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	// ReactTo (below) always produces exactly one child node on success and
	// never a partial/multi-stage progression, so "have I already reacted"
	// is exactly "does my own subflow already have a node" — flows.Nearest
	// here is this intent's own subflow (see NewSubFlow/Flows.Replace), not
	// any ancestor or sibling, so this cannot be satisfied by an unrelated
	// node elsewhere in the flow.
	if len(flows.Nearest.Nodes) > 0 {
		return nil, authflow.ErrEOF
	}

	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}
	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer, n)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeIdentificationOptionIndex{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}

func (n *IntentUseIdentitySelectAccount) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	// input is either a real client input carrying "index" (a direct call
	// against this login flow's own identify step), or a
	// SyntheticInputSelectAccount replayed from a signup_login switch,
	// carrying the user ID directly instead. n.ExpectedUserID is already
	// resolved by the caller either way — this check only gates that some
	// select_account-shaped input arrived, it doesn't need to extract
	// anything from it.
	var inputTakeIdentificationOptionIndex inputTakeIdentificationOptionIndex
	var inputTakeSelectAccountUserID inputTakeSelectAccountUserID
	if !authflow.AsInput(input, &inputTakeIdentificationOptionIndex) && !authflow.AsInput(input, &inputTakeSelectAccountUserID) {
		return nil, authflow.ErrIncompatibleInput
	}

	var bpSpecialErr error
	bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
	if err != nil {
		return nil, err
	}

	result, err := NewNodeDoUseIdentitySelectAccount(ctx, deps, flows, n.ExpectedUserID)
	if err != nil {
		return nil, err
	}
	return result, bpSpecialErr
}
