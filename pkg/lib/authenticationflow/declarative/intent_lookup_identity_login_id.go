package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

func init() {
	authflow.RegisterIntent(&IntentLookupIdentityLoginID{})
}

type IntentLookupIdentityLoginID struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	SyntheticInput *InputStepIdentify                      `json:"synthetic_input,omitempty"`
}

var _ authflow.Intent = &IntentLookupIdentityLoginID{}
var _ authflow.Milestone = &IntentLookupIdentityLoginID{}
var _ MilestoneIdentificationMethod = &IntentLookupIdentityLoginID{}
var _ authflow.InputReactor = &IntentLookupIdentityLoginID{}

func (*IntentLookupIdentityLoginID) Kind() string {
	return "IntentLookupIdentityLoginID"
}

func (*IntentLookupIdentityLoginID) Milestone() {}
func (n *IntentLookupIdentityLoginID) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *IntentLookupIdentityLoginID) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (n *IntentLookupIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)

	var inputTakeLoginID inputTakeLoginID

	if authflow.AsInput(input, &inputTakeLoginID) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}

		loginID := inputTakeLoginID.GetLoginID()
		spec := makeLoginIDSpec(n.Identification, stringutil.NewUserInputString(loginID))

		syntheticInput := &InputStepIdentify{
			Identification: n.SyntheticInput.Identification,
			LoginID:        loginID,
		}

		_, err = findExactOneIdentityInfo(ctx, deps, spec)
		if err != nil {
			if apierrors.IsKind(err, api.UserNotFound) {
				// signup
				return nil, errors.Join(bpSpecialErr, &authflow.ErrorSwitchFlow{
					FlowReference: authflow.FlowReference{
						Type: authflow.FlowTypeSignup,
						Name: oneOf.SignupFlow,
					},
					SyntheticInput: syntheticInput,
				})
			}
			// general error
			return nil, err
		}

		// login
		return nil, errors.Join(bpSpecialErr, &authflow.ErrorSwitchFlow{
			FlowReference: authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				Name: oneOf.LoginFlow,
			},
			SyntheticInput: syntheticInput,
		})
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentLookupIdentityLoginID) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}
