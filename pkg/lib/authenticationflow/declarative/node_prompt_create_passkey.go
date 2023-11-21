package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterNode(&NodePromptCreatePasskey{})
}

type nodePromptCreatePasskeyData struct {
	CreationOptions *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
}

var _ authflow.Data = &nodePromptCreatePasskeyData{}

func (m nodePromptCreatePasskeyData) Data() {}

type NodePromptCreatePasskey struct {
	JSONPointer     jsonpointer.T                  `json:"json_pointer,omitempty"`
	UserID          string                         `json:"user_id,omitempty"`
	CreationOptions *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
}

var _ authflow.NodeSimple = &NodePromptCreatePasskey{}
var _ authflow.InputReactor = &NodePromptCreatePasskey{}
var _ authflow.DataOutputer = &NodePromptCreatePasskey{}

func NewNodePromptCreatePasskey(deps *authflow.Dependencies, n *NodePromptCreatePasskey) (*NodePromptCreatePasskey, error) {
	creationOptions, err := deps.PasskeyCreationOptionsService.MakeCreationOptions(n.UserID)
	if err != nil {
		return nil, err
	}

	n.CreationOptions = creationOptions
	return n, nil
}

func (n *NodePromptCreatePasskey) Kind() string {
	return "NodePromptCreatePasskey"
}

func (n *NodePromptCreatePasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaPromptCreatePasskey{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodePromptCreatePasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputNodePromptCreatePasskey inputNodePromptCreatePasskey
	if !authflow.AsInput(input, &inputNodePromptCreatePasskey) {
		return nil, authflow.ErrIncompatibleInput
	}

	switch {
	case inputNodePromptCreatePasskey.IsCreationResponse():
		creationResponse := inputNodePromptCreatePasskey.GetCreationResponse()
		creationResponseBytes, err := json.Marshal(creationResponse)
		if err != nil {
			return nil, err
		}

		authenticatorSpec := &authenticator.Spec{
			UserID: n.UserID,
			Kind:   authenticator.KindPrimary,
			Type:   model.AuthenticatorTypePasskey,
			Passkey: &authenticator.PasskeySpec{
				AttestationResponse: creationResponseBytes,
			},
		}

		authenticatorID := uuid.New()
		authenticatorInfo, err := deps.Authenticators.NewWithAuthenticatorID(authenticatorID, authenticatorSpec)
		if err != nil {
			return nil, err
		}

		identitySpec := &identity.Spec{
			Type: model.IdentityTypePasskey,
			Passkey: &identity.PasskeySpec{
				AttestationResponse: creationResponseBytes,
			},
		}
		identityInfo, err := deps.Identities.New(n.UserID, identitySpec, identity.NewIdentityOptions{})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoCreatePasskey{
			Identity:            identityInfo,
			Authenticator:       authenticatorInfo,
			AttestationResponse: creationResponseBytes,
		}), nil
	case inputNodePromptCreatePasskey.IsSkip():
		return authflow.NewNodeSimple(&NodeSentinel{}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (n *NodePromptCreatePasskey) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return nodePromptCreatePasskeyData{
		CreationOptions: n.CreationOptions,
	}, nil
}
