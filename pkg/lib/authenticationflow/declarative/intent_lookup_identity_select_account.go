package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLookupIdentitySelectAccount{})
}

type IntentLookupIdentitySelectAccount struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	SyntheticInput *InputStepIdentify                     `json:"synthetic_input,omitempty"`
	ExpectedUserID string                                 `json:"expected_user_id,omitempty"`
}

var _ authflow.Intent = &IntentLookupIdentitySelectAccount{}
var _ authflow.Milestone = &IntentLookupIdentitySelectAccount{}
var _ MilestoneIdentificationMethod = &IntentLookupIdentitySelectAccount{}
var _ authflow.InputReactor = &IntentLookupIdentitySelectAccount{}

func (*IntentLookupIdentitySelectAccount) Kind() string {
	return "IntentLookupIdentitySelectAccount"
}

func (*IntentLookupIdentitySelectAccount) Milestone() {}

func (n *IntentLookupIdentitySelectAccount) MilestoneIdentificationMethod() model.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *IntentLookupIdentitySelectAccount) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
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

func (n *IntentLookupIdentitySelectAccount) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	flowRootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}
	oneOf := n.oneOf(current)

	var inputTakeIdentificationOptionIndex inputTakeIdentificationOptionIndex
	if authflow.AsInput(input, &inputTakeIdentificationOptionIndex) {
		var bpSpecialErr error
		bpSpecialErr, err := HandleBotProtection(ctx, deps, flows, n.JSONPointer, input, n)
		if err != nil {
			return nil, err
		}

		if _, err := resolveSelectAccountSession(ctx, n.ExpectedUserID); err != nil {
			return nil, err
		}

		// select_account never switches to signup: it can only ever
		// continue an existing login.
		return nil, errors.Join(bpSpecialErr, &authflow.ErrorSwitchFlow{
			FlowReference: authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				Name: oneOf.LoginFlow,
			},
			SyntheticInput: n.SyntheticInput,
		})
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentLookupIdentitySelectAccount) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}
