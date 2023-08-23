package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterNode(&NodeCreateAuthenticatorTOTP{})
}

type NodeCreateAuthenticatorTOTPData struct {
	Secret string `json:"secret"`
}

var _ authflow.Data = NodeCreateAuthenticatorTOTPData{}

func (m NodeCreateAuthenticatorTOTPData) Data() {}

type NodeCreateAuthenticatorTOTP struct {
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	Authenticator  *authenticator.Info                     `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeCreateAuthenticatorTOTP{}
var _ authflow.Milestone = &NodeCreateAuthenticatorTOTP{}
var _ MilestoneAuthenticationMethod = &NodeCreateAuthenticatorTOTP{}
var _ authflow.InputReactor = &NodeCreateAuthenticatorTOTP{}
var _ authflow.DataOutputer = &NodeCreateAuthenticatorTOTP{}

func NewNodeCreateAuthenticatorTOTP(deps *authflow.Dependencies, n *NodeCreateAuthenticatorTOTP) (*NodeCreateAuthenticatorTOTP, error) {
	authenticatorKind := n.authenticatorKind()

	isDefault, err := authenticatorIsDefault(deps, n.UserID, authenticatorKind)
	if err != nil {
		return nil, err
	}

	spec := &authenticator.Spec{
		UserID:    n.UserID,
		IsDefault: isDefault,
		Kind:      authenticatorKind,
		Type:      model.AuthenticatorTypeTOTP,
		TOTP: &authenticator.TOTPSpec{
			// The display name will be filled by input.
			DisplayName: "",
		},
	}

	id := uuid.New()
	info, err := deps.Authenticators.NewWithAuthenticatorID(id, spec)
	if err != nil {
		return nil, err
	}

	n.Authenticator = info
	return n, nil
}

func (*NodeCreateAuthenticatorTOTP) Kind() string {
	return "NodeCreateAuthenticatorTOTP"
}

func (*NodeCreateAuthenticatorTOTP) Milestone() {}
func (n *NodeCreateAuthenticatorTOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*NodeCreateAuthenticatorTOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSetupTOTP{}, nil
}

func (i *NodeCreateAuthenticatorTOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputSetupTOTP inputSetupTOTP
	if authflow.AsInput(input, &inputSetupTOTP) {
		_, err := deps.Authenticators.VerifyWithSpec(i.Authenticator, &authenticator.Spec{
			TOTP: &authenticator.TOTPSpec{
				Code: inputSetupTOTP.GetCode(),
			},
		}, nil)
		if errors.Is(err, api.ErrInvalidCredentials) {
			return nil, api.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}

		// Set display name.
		i.Authenticator.TOTP.DisplayName = inputSetupTOTP.GetDisplayName()
		return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: i.Authenticator,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeCreateAuthenticatorTOTP) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	secret := n.Authenticator.TOTP.Secret
	return NodeCreateAuthenticatorTOTPData{
		Secret: secret,
	}, nil
}

func (n *NodeCreateAuthenticatorTOTP) authenticatorKind() model.AuthenticatorKind {
	switch n.Authentication {
	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		return model.AuthenticatorKindSecondary
	default:
		panic(fmt.Errorf("unexpected authentication method: %v", n.Authentication))
	}
}
