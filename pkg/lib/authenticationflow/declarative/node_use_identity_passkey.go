package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterNode(&NodeUseIdentityPasskey{})
}

type NodeUseIdentityPasskey struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseIdentityPasskey{}
var _ authflow.Milestone = &NodeUseIdentityPasskey{}
var _ MilestoneIdentificationMethod = &NodeUseIdentityPasskey{}
var _ authflow.InputReactor = &NodeUseIdentityPasskey{}

func (*NodeUseIdentityPasskey) Kind() string {
	return "NodeUseIdentityPasskey"
}

func (*NodeUseIdentityPasskey) Milestone() {}
func (n *NodeUseIdentityPasskey) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *NodeUseIdentityPasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakePasskeyAssertionResponse{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *NodeUseIdentityPasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputAssertionResponse inputTakePasskeyAssertionResponse
	if authflow.AsInput(input, &inputAssertionResponse) {
		assertionResponse := inputAssertionResponse.GetAssertionResponse()
		assertionResponseBytes, err := json.Marshal(assertionResponse)
		if err != nil {
			return nil, err
		}

		identitySpec := &identity.Spec{
			Type: model.IdentityTypePasskey,
			Passkey: &identity.PasskeySpec{
				AssertionResponse: assertionResponseBytes,
			},
		}

		exactMatch, err := findExactOneIdentityInfo(deps, identitySpec)
		if err != nil {
			return nil, err
		}

		userID := exactMatch.UserID

		authenticatorSpec := &authenticator.Spec{
			Type: model.AuthenticatorTypePasskey,
			Passkey: &authenticator.PasskeySpec{
				AssertionResponse: assertionResponseBytes,
			},
		}

		authenticators, err := deps.Authenticators.List(userID, authenticator.KeepType(model.AuthenticatorTypePasskey))
		if err != nil {
			return nil, err
		}

		authenticatorInfo, verifyResult, err := deps.Authenticators.VerifyOneWithSpec(
			userID,
			model.AuthenticatorTypePasskey,
			authenticators,
			authenticatorSpec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					userID,
					authn.AuthenticationStagePrimary,
					authn.AuthenticationTypePasskey,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		n, err := NewNodeDoUseIdentityPasskey(ctx, flows, &NodeDoUseIdentityPasskey{
			AssertionResponse: assertionResponseBytes,
			Identity:          exactMatch,
			Authenticator:     authenticatorInfo,
			RequireUpdate:     verifyResult.Passkey,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
