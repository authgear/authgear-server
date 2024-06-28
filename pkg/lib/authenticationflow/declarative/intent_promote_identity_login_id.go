package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
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
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakeLoginID{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *IntentPromoteIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		loginID := inputTakeLoginID.GetLoginID()
		specForLookup := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Value: loginID,
			},
		}

		syntheticInput := &InputStepIdentify{
			Identification: n.SyntheticInput.Identification,
			LoginID:        loginID,
		}

		_, err := findExactOneIdentityInfo(deps, specForLookup)
		if err != nil {
			// promote
			if apierrors.IsKind(err, api.UserNotFound) {
				spec := n.makeLoginIDSpec(loginID)

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
					return nil, identityFillDetailsMany(api.ErrDuplicatedIdentity, spec, conflictSpecs)
				}

				info, err := newIdentityInfo(deps, n.UserID, spec)
				if err != nil {
					return nil, err
				}

				return authflow.NewNodeSimple(&NodeDoCreateIdentity{
					Identity: info,
				}), nil

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

func (n *IntentPromoteIdentityLoginID) makeLoginIDSpec(loginID string) *identity.Spec {
	spec := &identity.Spec{
		Type: model.IdentityTypeLoginID,
		LoginID: &identity.LoginIDSpec{
			Value: loginID,
		},
	}
	switch n.Identification {
	case config.AuthenticationFlowIdentificationEmail:
		spec.LoginID.Type = model.LoginIDKeyTypeEmail
		spec.LoginID.Key = string(spec.LoginID.Type)
	case config.AuthenticationFlowIdentificationPhone:
		spec.LoginID.Type = model.LoginIDKeyTypePhone
		spec.LoginID.Key = string(spec.LoginID.Type)
	case config.AuthenticationFlowIdentificationUsername:
		spec.LoginID.Type = model.LoginIDKeyTypeUsername
		spec.LoginID.Key = string(spec.LoginID.Type)
	default:
		panic(fmt.Errorf("unexpected identification method: %v", n.Identification))
	}

	return spec
}

func (n *IntentPromoteIdentityLoginID) checkConflictByAccountLinkings(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	spec *identity.Spec) (conflicts []*AccountLinkingConflict, err error) {
	switch spec.Type {
	case model.IdentityTypeLoginID:
		return linkByIncomingLoginIDSpec(ctx, deps, flows, NewCreateLoginIDIdentityRequest(spec).LoginID, n.JSONPointer)
	default:
		panic("unexpected spec type")
	}
}
