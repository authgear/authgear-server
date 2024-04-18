package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterNode(&NodeCreateAuthenticatorPassword{})
}

type NodeCreateAuthenticatorPassword struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeCreateAuthenticatorPassword{}
var _ authflow.InputReactor = &NodeCreateAuthenticatorPassword{}
var _ authflow.Milestone = &NodeCreateAuthenticatorPassword{}
var _ MilestoneAuthenticationMethod = &NodeCreateAuthenticatorPassword{}
var _ MilestoneSwitchToExistingUser = &NodeCreateAuthenticatorPassword{}

func (*NodeCreateAuthenticatorPassword) Kind() string {
	return "NodeCreateAuthenticatorPassword"
}

func (*NodeCreateAuthenticatorPassword) Milestone() {}
func (n *NodeCreateAuthenticatorPassword) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}
func (i *NodeCreateAuthenticatorPassword) MilestoneSwitchToExistingUser(newUserID string) {
	// TODO(tung): Skip creation if already have one
	i.UserID = newUserID
}

func (n *NodeCreateAuthenticatorPassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeNewPassword{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (i *NodeCreateAuthenticatorPassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if authflow.AsInput(input, &inputTakeNewPassword) {
		authenticatorKind := i.Authentication.AuthenticatorKind()
		newPassword := inputTakeNewPassword.GetNewPassword()
		isDefault, err := authenticatorIsDefault(deps, i.UserID, authenticatorKind)
		if err != nil {
			return nil, err
		}

		spec := &authenticator.Spec{
			UserID:    i.UserID,
			IsDefault: isDefault,
			Kind:      authenticatorKind,
			Type:      model.AuthenticatorTypePassword,
			Password: &authenticator.PasswordSpec{
				PlainPassword: newPassword,
			},
		}

		authenticatorID := uuid.New()
		info, err := deps.Authenticators.NewWithAuthenticatorID(authenticatorID, spec)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
