package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterNode(&NodeUseAuthenticatorPasskey{})
}

type NodeUseAuthenticatorPasskey struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAuthenticatorPasskey{}
var _ authflow.Milestone = &NodeUseAuthenticatorPasskey{}
var _ MilestoneAuthenticationMethod = &NodeUseAuthenticatorPasskey{}
var _ MilestoneFlowAuthenticate = &NodeUseAuthenticatorPasskey{}
var _ authflow.InputReactor = &NodeUseAuthenticatorPasskey{}

func (*NodeUseAuthenticatorPasskey) Kind() string {
	return "NodeUseAuthenticatorPasskey"
}

func (*NodeUseAuthenticatorPasskey) Milestone() {}
func (n *NodeUseAuthenticatorPasskey) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*NodeUseAuthenticatorPasskey) MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
}

func (n *NodeUseAuthenticatorPasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakePasskeyAssertionResponse{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *NodeUseAuthenticatorPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputAssertionResponse inputTakePasskeyAssertionResponse
	if authflow.AsInput(input, &inputAssertionResponse) {
		assertionResponse := inputAssertionResponse.GetAssertionResponse()
		assertionResponseBytes, err := json.Marshal(assertionResponse)
		if err != nil {
			return nil, err
		}

		authenticatorSpec := &authenticator.Spec{
			Type: model.AuthenticatorTypePasskey,
			Passkey: &authenticator.PasskeySpec{
				AssertionResponse: assertionResponseBytes,
			},
		}

		authenticators, err := deps.Authenticators.List(n.UserID, authenticator.KeepType(model.AuthenticatorTypePasskey))
		if err != nil {
			return nil, err
		}

		authenticatorInfo, verifyResult, err := deps.Authenticators.VerifyOneWithSpec(
			n.UserID,
			model.AuthenticatorTypePasskey,
			authenticators,
			authenticatorSpec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					n.UserID,
					authn.AuthenticationStagePrimary,
					authn.AuthenticationTypePasskey,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoUseAuthenticatorPasskey{
			AssertionResponse: assertionResponseBytes,
			Authenticator:     authenticatorInfo,
			RequireUpdate:     verifyResult.Passkey,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
