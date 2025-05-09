package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLookupWithIDToken{})
}

type IntentLookupWithIDToken struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	SyntheticInput *InputStepIdentify                      `json:"synthetic_input,omitempty"`
}

var _ authflow.Intent = &IntentLookupWithIDToken{}
var _ authflow.Milestone = &IntentLookupWithIDToken{}
var _ MilestoneIdentificationMethod = &IntentLookupWithIDToken{}
var _ authflow.InputReactor = &IntentLookupWithIDToken{}

func (*IntentLookupWithIDToken) Kind() string {
	return "IntentLookupWithIDToken"
}

func (*IntentLookupWithIDToken) Milestone() {}
func (n *IntentLookupWithIDToken) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *IntentLookupWithIDToken) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeIDToken{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *IntentLookupWithIDToken) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)

	var inputTakeIDToken inputTakeIDToken

	if authflow.AsInput(input, &inputTakeIDToken) {
		idToken := inputTakeIDToken.GetIDToken()

		syntheticInput := &InputStepIdentify{
			Identification: n.SyntheticInput.Identification,
			IDToken:        idToken,
		}

		token, err := deps.IDTokens.VerifyIDToken(idToken)
		if err != nil {
			return nil, apierrors.NewInvalid("invalid ID token")
		}

		userID := token.Subject()
		_, err = deps.Users.GetRaw(ctx, userID)
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				return nil, api.ErrUserNotFound
			}

			// When ID token is used, we never switch to signup flow.
			return nil, err
		}

		// Switch to login flow.
		return nil, &authflow.ErrorSwitchFlow{
			FlowReference: authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				Name: oneOf.LoginFlow,
			},
			SyntheticInput: syntheticInput,
		}
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentLookupWithIDToken) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}
