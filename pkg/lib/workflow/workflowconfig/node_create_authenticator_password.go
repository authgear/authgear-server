package workflowconfig

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	workflow.RegisterNode(&NodeCreateAuthenticatorPassword{})
}

type NodeCreateAuthenticatorPassword struct {
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

func (*NodeCreateAuthenticatorPassword) Kind() string {
	return "workflowconfig.NodeCreateAuthenticatorPassword"
}

func (*NodeCreateAuthenticatorPassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeNewPassword{}}, nil
}

func (i *NodeCreateAuthenticatorPassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if workflow.AsInput(input, &inputTakeNewPassword) {
		authenticatorKind := i.authenticatorKind()
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

		return workflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*NodeCreateAuthenticatorPassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeCreateAuthenticatorPassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *NodeCreateAuthenticatorPassword) authenticatorKind() model.AuthenticatorKind {
	switch i.Authentication {
	case config.WorkflowAuthenticationMethodPrimaryPassword:
		return model.AuthenticatorKindPrimary
	case config.WorkflowAuthenticationMethodSecondaryPassword:
		return model.AuthenticatorKindSecondary
	default:
		panic(fmt.Errorf("workflow: unexpected authentication method: %v", i.Authentication))
	}
}
