package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

func init() {
	authflow.RegisterIntent(&IntentPromoteIdentityLoginID{})
}

type IntentPromoteIdentityLoginID struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	SyntheticInput *InputStepIdentify                      `json:"synthetic_input,omitempty"`
}

var _ authflow.Intent = &IntentPromoteIdentityLoginID{}
var _ authflow.Milestone = &IntentPromoteIdentityLoginID{}
var _ MilestoneIdentificationMethod = &IntentPromoteIdentityLoginID{}
var _ MilestoneFlowCreateIdentity = &IntentPromoteIdentityLoginID{}
var _ authflow.InputReactor = &IntentPromoteIdentityLoginID{}

func (*IntentPromoteIdentityLoginID) Kind() string {
	return "IntentPromoteIdentityLoginID"
}

func (*IntentPromoteIdentityLoginID) Milestone() {}
func (n *IntentPromoteIdentityLoginID) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}
func (n *IntentPromoteIdentityLoginID) MilestoneFlowCreateIdentity(flows authflow.Flows) (MilestoneDoCreateIdentity, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateIdentity](flows)
}

func (n *IntentPromoteIdentityLoginID) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, identified := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateIdentity](flows)
	if identified {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
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

// nolint:gocognit
func (n *IntentPromoteIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input)
		if err != nil {
			return nil, err
		}
		loginID := inputTakeLoginID.GetLoginID()
		specForLookup := makeLoginIDSpec(n.Identification, stringutil.NewUserInputString(loginID))

		syntheticInput := &InputStepIdentify{
			Identification: n.SyntheticInput.Identification,
			LoginID:        loginID,
		}

		_, err = findExactOneIdentityInfo(ctx, deps, specForLookup)
		if err != nil {
			// promote
			if apierrors.IsKind(err, api.UserNotFound) {
				spec := makeLoginIDSpec(n.Identification, stringutil.NewUserInputString(loginID))

				conflicts, err := n.checkConflictByAccountLinkings(ctx, deps, flows, spec)
				if err != nil {
					return nil, err
				}
				if len(conflicts) > 0 {
					// In promote flow, always error if any conflicts occurs
					conflictSpecs := slice.Map(conflicts, func(c *AccountLinkingConflict) *identity.Spec {
						s := c.Identity.ToSpec()
						return &s
					})
					return nil, identity.NewErrDuplicatedIdentityMany(spec, conflictSpecs)
				}

				info, err := newIdentityInfo(ctx, deps, n.UserID, spec)
				if err != nil {
					return nil, err
				}

				reactToResult, err := NewNodeDoCreateIdentityReactToResult(ctx, flows, deps, NodeDoCreateIdentityOptions{
					SkipCreate:   false,
					Identity:     info,
					IdentitySpec: spec,
				})
				if err != nil {
					return nil, err
				}

				return reactToResult, bpSpecialErr

			}
			// general error
			return nil, err
		}

		// login
		flowReference := authflow.FindCurrentFlowReference(flows.Root)
		return nil, &authflow.ErrorSwitchFlow{
			FlowReference: authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				// Switch to the login flow of the same name.
				Name: flowReference.Name,
			},
			SyntheticInput: syntheticInput,
		}
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentPromoteIdentityLoginID) checkConflictByAccountLinkings(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	spec *identity.Spec) (conflicts []*AccountLinkingConflict, err error) {
	switch spec.Type {
	case model.IdentityTypeLoginID:
		return linkByIncomingLoginIDSpec(ctx, deps, flows, n.UserID, NewCreateLoginIDIdentityRequest(spec).LoginID, n.JSONPointer)
	default:
		panic("unexpected spec type")
	}
}
