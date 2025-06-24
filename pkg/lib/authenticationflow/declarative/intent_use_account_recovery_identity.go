package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
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
	_, _, recovered := authflow.FindMilestoneInCurrentFlow[MilestoneDoUseAccountRecoveryIdentity](flows)
	if recovered {
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
	return &InputSchemaTakeLoginID{
		FlowRootObject:          flowRootObject,
		JSONPointer:             n.JSONPointer,
		IsBotProtectionRequired: isBotProtectionRequired,
		BotProtectionCfg:        deps.Config.BotProtection,
	}, nil
}

func (n *IntentUseAccountRecoveryIdentity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}
		loginID := inputTakeLoginID.GetLoginID()
		spec := makeLoginIDSpec(n.Identification.AuthenticationFlowIdentification(), stringutil.NewUserInputString(loginID))

		var exactMatch *identity.Info = nil
		switch n.OnFailure {
		case config.AuthenticationFlowAccountRecoveryIdentificationOnFailureIgnore:
			var err error
			exactMatch, _, err = deps.Identities.SearchBySpec(ctx, spec)
			if err != nil {
				return nil, err
			}
		case config.AuthenticationFlowAccountRecoveryIdentificationOnFailureError:
			var err error
			// This reservation cannot be adjusted by hook,
			// because account_recovery flow does not trigget authentication.post_identified,
			// so ignore it
			exactMatch, _, err = findExactOneIdentityInfo(ctx, deps, spec)
			if err != nil {
				return nil, err
			}
		}

		nextNode := &NodeDoUseAccountRecoveryIdentity{
			Identification: n.Identification,
			Spec:           spec,
			MaybeIdentity:  exactMatch,
		}

		return authflow.NewNodeSimple(nextNode), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
