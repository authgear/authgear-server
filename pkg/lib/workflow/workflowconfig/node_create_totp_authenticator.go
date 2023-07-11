package workflowconfig

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	workflow.RegisterNode(&NodeCreateTOTPAuthenticator{})
}

type NodeCreateTOTPAuthenticator struct {
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
	Authenticator  *authenticator.Info                 `json:"authenticator,omitempty"`
}

func NewNodeCreateTOTPAuthenticator(deps *workflow.Dependencies, n *NodeCreateTOTPAuthenticator) (*NodeCreateTOTPAuthenticator, error) {
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

func (*NodeCreateTOTPAuthenticator) Kind() string {
	return "workflowconfig.NodeCreateTOTPAuthenticator"
}

func (*NodeCreateTOTPAuthenticator) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputSetupTOTP{}}, nil
}

func (i *NodeCreateTOTPAuthenticator) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputSetupTOTP inputSetupTOTP
	if workflow.AsInput(input, &inputSetupTOTP) {
		_, err := deps.Authenticators.VerifyWithSpec(i.Authenticator, &authenticator.Spec{
			TOTP: &authenticator.TOTPSpec{
				Code: inputSetupTOTP.GetCode(),
			},
		})
		if errors.Is(err, authenticator.ErrInvalidCredentials) {
			return nil, api.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}

		// Set display name.
		i.Authenticator.TOTP.DisplayName = inputSetupTOTP.GetDisplayName()
		return workflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: i.Authenticator,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*NodeCreateTOTPAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeCreateTOTPAuthenticator) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	secret := n.Authenticator.TOTP.Secret
	return map[string]interface{}{
		"secret": secret,
	}, nil
}

func (n *NodeCreateTOTPAuthenticator) authenticatorKind() model.AuthenticatorKind {
	switch n.Authentication {
	case config.WorkflowAuthenticationMethodSecondaryTOTP:
		return model.AuthenticatorKindSecondary
	default:
		panic(fmt.Errorf("workflow: unexpected authentication method: %v", n.Authentication))
	}
}
