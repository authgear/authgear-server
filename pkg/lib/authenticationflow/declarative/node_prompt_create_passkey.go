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

type NodePromptCreatePasskeyData struct {
	TypedData
	CreationOptions *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
}

func NewNodePromptCreatePasskeyData(d NodePromptCreatePasskeyData) NodePromptCreatePasskeyData {
	d.Type = DataTypeCreatePasskeyData
	return d
}

var _ authflow.Data = &NodePromptCreatePasskeyData{}

func (m NodePromptCreatePasskeyData) Data() {}

type NodePromptCreatePasskey struct {
	JSONPointer     jsonpointer.T                  `json:"json_pointer,omitempty"`
	UserID          string                         `json:"user_id,omitempty"`
	CreationOptions *model.WebAuthnCreationOptions `json:"creation_options,omitempty"`
}

var _ authflow.NodeSimple = &NodePromptCreatePasskey{}
var _ authflow.InputReactor = &NodePromptCreatePasskey{}
var _ authflow.DataOutputer = &NodePromptCreatePasskey{}
var _ authflow.Milestone = &NodePromptCreatePasskey{}
var _ MilestonePromptCreatePasskey = &NodePromptCreatePasskey{}

func NewNodePromptCreatePasskey(ctx context.Context, deps *authflow.Dependencies, n *NodePromptCreatePasskey) (*NodePromptCreatePasskey, error) {
	creationOptions, err := deps.PasskeyCreationOptionsService.MakeCreationOptions(ctx, n.UserID)
	if err != nil {
		return nil, err
	}

	n.CreationOptions = creationOptions
	return n, nil
}

func (n *NodePromptCreatePasskey) Kind() string {
	return "NodePromptCreatePasskey"
}

func (n *NodePromptCreatePasskey) Milestone()                    {}
func (n *NodePromptCreatePasskey) MilestonePromptCreatePasskey() {}

func (n *NodePromptCreatePasskey) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {

	if n.isAlreadyPrompted(flows) {
		// Don't ask for input if already prompted once
		return nil, nil
	}

	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaPromptCreatePasskey{
		JSONPointer:    n.JSONPointer,
		FlowRootObject: flowRootObject,
	}, nil
}

func (n *NodePromptCreatePasskey) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	if n.isAlreadyPrompted(flows) {
		return authflow.NewNodeSimple(&NodeSentinel{}), nil
	}

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
		authenticatorInfo, err := deps.Authenticators.NewWithAuthenticatorID(ctx, authenticatorID, authenticatorSpec)
		if err != nil {
			return nil, err
		}

		identitySpec := &identity.Spec{
			Type: model.IdentityTypePasskey,
			Passkey: &identity.PasskeySpec{
				AttestationResponse: creationResponseBytes,
			},
		}
		identityInfo, err := deps.Identities.New(ctx, n.UserID, identitySpec, identity.NewIdentityOptions{})
		if err != nil {
			return nil, err
		}

		return NewNodeDoCreatePasskeyReactToResult(ctx, deps, NodeDoCreatePasskeyOptions{
			Identity:            identityInfo,
			Authenticator:       authenticatorInfo,
			AttestationResponse: creationResponseBytes,
		})
	case inputNodePromptCreatePasskey.IsSkip():
		return authflow.NewNodeSimple(&NodeSentinel{}), nil
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (n *NodePromptCreatePasskey) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	return NewNodePromptCreatePasskeyData(NodePromptCreatePasskeyData{
		CreationOptions: n.CreationOptions,
	}), nil
}

func (n *NodePromptCreatePasskey) isAlreadyPrompted(flows authflow.Flows) bool {
	ms := authflow.FindAllMilestones[MilestonePromptCreatePasskey](flows.Root)

	if len(ms) == 0 {
		return false
	}

	for _, m := range ms {
		// Another milestone was found => already prompted.
		if m != n {
			return true
		}
	}

	// Otherwise len(ms) > 0 and all milestone == n, => not prompted yet.
	return false
}
