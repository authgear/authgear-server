package declarative

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentLookupIdentityPasskey{})
}

type IntentLookupIdentityPasskey struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	SyntheticInput *InputStepIdentify                      `json:"synthetic_input,omitempty"`
}

var _ authflow.Intent = &IntentLookupIdentityPasskey{}
var _ authflow.Milestone = &IntentLookupIdentityPasskey{}
var _ MilestoneIdentificationMethod = &IntentLookupIdentityPasskey{}
var _ authflow.InputReactor = &IntentLookupIdentityPasskey{}

func (*IntentLookupIdentityPasskey) Kind() string {
	return "IntentLookupIdentityPasskey"
}

func (*IntentLookupIdentityPasskey) Milestone() {}
func (n *IntentLookupIdentityPasskey) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *IntentLookupIdentityPasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakePasskeyAssertionResponse{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *IntentLookupIdentityPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)

	var inputAssertionResponse inputTakePasskeyAssertionResponse
	if authflow.AsInput(input, &inputAssertionResponse) {
		assertionResponse := inputAssertionResponse.GetAssertionResponse()
		assertionResponseBytes, err := json.Marshal(assertionResponse)
		if err != nil {
			return nil, err
		}

		syntheticInput := &SyntheticInputPasskey{
			Identification:    n.SyntheticInput.Identification,
			AssertionResponse: assertionResponse,
		}

		spec := &identity.Spec{
			Type: model.IdentityTypePasskey,
			Passkey: &identity.PasskeySpec{
				AssertionResponse: assertionResponseBytes,
			},
		}

		_, err = findExactOneIdentityInfo(deps, spec)
		if err != nil {
			if apierrors.IsKind(err, api.UserNotFound) {
				// signup
				// We do not support sign up with passkey.
				return nil, err
			}
			// general error
			return nil, err
		}

		// login
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

func (n *IntentLookupIdentityPasskey) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}
