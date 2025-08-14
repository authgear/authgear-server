package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

func init() {
	authflow.RegisterIntent(&IntentCreateIdentityLoginID{})
}

type IntentCreateIdentityLoginID struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	UserID         string                                 `json:"user_id,omitempty"`
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`
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

func (n *IntentCreateIdentityLoginID) MilestoneIdentificationMethod() model.AuthenticationFlowIdentification {
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

func (n *IntentCreateIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeLoginIDOrExternalJWT inputTakeLoginIDOrExternalJWT
	if authflow.AsInput(input, &inputTakeLoginIDOrExternalJWT) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}

		externalJWT := inputTakeLoginIDOrExternalJWT.GetExternalJWT()
		if externalJWT != "" {
			verifiedToken, err := deps.ExternalJWT.VerifyExternalJWT(ctx, externalJWT)
			if err != nil {
				return nil, err
			}

			spec, err := deps.ExternalJWT.ConstructLoginIDSpec(
				n.Identification,
				verifiedToken,
			)
			if err != nil {
				return nil, err
			}

			return authflow.NewSubFlow(&IntentCheckConflictAndCreateIdenity{
				JSONPointer: n.JSONPointer,
				UserID:      n.UserID,
				Request:     NewCreateLoginIDIdentityRequest(spec),
			}), bpSpecialErr
		}

		loginID := inputTakeLoginIDOrExternalJWT.GetLoginID()
		spec := makeLoginIDSpec(n.Identification, stringutil.NewUserInputString(loginID))

		return authflow.NewSubFlow(&IntentCheckConflictAndCreateIdenity{
			JSONPointer: n.JSONPointer,
			UserID:      n.UserID,
			Request:     NewCreateLoginIDIdentityRequest(spec),
		}), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}
