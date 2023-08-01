package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeUseAuthenticatorPassword{})
}

type NodeUseAuthenticatorPassword struct {
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

func (*NodeUseAuthenticatorPassword) Kind() string {
	return "workflowconfig.NodeUseAuthenticatorPassword"
}

func (*NodeUseAuthenticatorPassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakePassword{}}, nil
}

func (i *NodeUseAuthenticatorPassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakePassword inputTakePassword
	if workflow.AsInput(input, &inputTakePassword) {
		as, err := deps.Authenticators.List(
			i.UserID,
			authenticator.KeepKind(i.Authentication.AuthenticatorKind()),
			authenticator.KeepType(model.AuthenticatorTypePassword),
		)
		if err != nil {
			return nil, err
		}

		password := inputTakePassword.GetPassword()
		spec := &authenticator.Spec{
			Password: &authenticator.PasswordSpec{
				PlainPassword: password,
			},
		}

		// FIXME(workflow): handle password forced change.
		info, _, err := deps.Authenticators.VerifyOneWithSpec(
			as,
			spec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					i.UserID,
					authn.AuthenticationStageFromAuthenticationMethod(i.Authentication),
					authn.AuthenticationTypePassword,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeDoUseAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*NodeUseAuthenticatorPassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeUseAuthenticatorPassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
