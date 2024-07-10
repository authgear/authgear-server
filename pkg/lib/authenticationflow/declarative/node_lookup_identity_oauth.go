package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeLookupIdentityOAuth{})
}

type NodeLookupIdentityOAuth struct {
	JSONPointer    jsonpointer.T      `json:"json_pointer,omitempty"`
	SyntheticInput *InputStepIdentify `json:"synthetic_input,omitempty"`
	Alias          string             `json:"alias,omitempty"`
	RedirectURI    string             `json:"redirect_uri,omitempty"`
	ResponseMode   string             `json:"response_mode,omitempty"`
}

var _ authflow.NodeSimple = &NodeLookupIdentityOAuth{}
var _ authflow.InputReactor = &NodeLookupIdentityOAuth{}
var _ authflow.DataOutputer = &NodeLookupIdentityOAuth{}

func (*NodeLookupIdentityOAuth) Kind() string {
	return "NodeLookupIdentityOAuth"
}

func (n *NodeLookupIdentityOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakeOAuthAuthorizationResponse{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *NodeLookupIdentityOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	current, err := authflow.FlowObject(flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}

	oneOf := n.oneOf(current)

	var inputOAuth inputTakeOAuthAuthorizationResponse
	if authflow.AsInput(input, &inputOAuth) {
		spec, err := handleOAuthAuthorizationResponse(deps, HandleOAuthAuthorizationResponseOptions{
			Alias:       n.Alias,
			RedirectURI: n.RedirectURI,
		}, inputOAuth)
		if err != nil {
			return nil, err
		}

		syntheticInput := &SyntheticInputOAuth{
			Identification: n.SyntheticInput.Identification,
			Alias:          n.SyntheticInput.Alias,
			RedirectURI:    n.SyntheticInput.RedirectURI,
			ResponseMode:   n.SyntheticInput.ResponseMode,
			IdentitySpec:   spec,
			BotProtection:  n.SyntheticInput.BotProtection,
		}

		_, err = findExactOneIdentityInfo(deps, spec)
		if err != nil {
			if apierrors.IsKind(err, api.UserNotFound) {
				// signup
				return nil, &authflow.ErrorSwitchFlow{
					FlowReference: authflow.FlowReference{
						Type: authflow.FlowTypeSignup,
						Name: oneOf.SignupFlow,
					},
					SyntheticInput: syntheticInput,
				}
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

func (n *NodeLookupIdentityOAuth) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	data, err := getOAuthData(ctx, deps, GetOAuthDataOptions{
		RedirectURI:  n.RedirectURI,
		Alias:        n.Alias,
		ResponseMode: n.ResponseMode,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (n *NodeLookupIdentityOAuth) oneOf(o config.AuthenticationFlowObject) *config.AuthenticationFlowSignupLoginFlowOneOf {
	oneOf, ok := o.(*config.AuthenticationFlowSignupLoginFlowOneOf)
	if !ok {
		panic(fmt.Errorf("flow object is %T", o))
	}

	return oneOf
}
