package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	workflow.RegisterNode(&NodeCreateAuthenticatorPassword{})
}

type NodeCreateAuthenticatorPassword struct {
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

var _ workflow.NodeSimple = &NodeCreateAuthenticatorPassword{}
var _ workflow.InputReactor = &NodeCreateAuthenticatorPassword{}
var _ workflow.Milestone = &NodeCreateAuthenticatorPassword{}
var _ MilestoneAuthenticationMethod = &NodeCreateAuthenticatorPassword{}

func (*NodeCreateAuthenticatorPassword) Kind() string {
	return "workflowconfig.NodeCreateAuthenticatorPassword"
}

func (*NodeCreateAuthenticatorPassword) Milestone() {}
func (n *NodeCreateAuthenticatorPassword) MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return n.Authentication
}

func (*NodeCreateAuthenticatorPassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	return &InputTakeNewPassword{}, nil
}

func (i *NodeCreateAuthenticatorPassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeNewPassword inputTakeNewPassword
	if workflow.AsInput(input, &inputTakeNewPassword) {
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

		return workflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}
