package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeUseAuthenticatorPassword{})
}

type NodeUseAuthenticatorPassword struct {
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

var _ workflow.NodeSimple = &NodeUseAuthenticatorPassword{}
var _ workflow.Milestone = &NodeUseAuthenticatorPassword{}
var _ MilestoneAuthenticationMethod = &NodeUseAuthenticatorPassword{}
var _ workflow.InputReactor = &NodeUseAuthenticatorPassword{}

func (*NodeUseAuthenticatorPassword) Kind() string {
	return "workflowconfig.NodeUseAuthenticatorPassword"
}

func (*NodeUseAuthenticatorPassword) Milestone() {}
func (n *NodeUseAuthenticatorPassword) MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return n.Authentication
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

		info, requireUpdate, err := deps.Authenticators.VerifyOneWithSpec(
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

		return workflow.NewNodeSimple(&NodeDidVerifyAuthenticator{
			Authenticator:          info,
			PasswordChangeRequired: requireUpdate,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}
