package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterNode(&NodeUseAuthenticatorPassword{})
}

type NodeUseAuthenticatorPassword struct {
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAuthenticatorPassword{}
var _ authflow.Milestone = &NodeUseAuthenticatorPassword{}
var _ MilestoneAuthenticationMethod = &NodeUseAuthenticatorPassword{}
var _ authflow.InputReactor = &NodeUseAuthenticatorPassword{}

func (*NodeUseAuthenticatorPassword) Kind() string {
	return "NodeUseAuthenticatorPassword"
}

func (*NodeUseAuthenticatorPassword) Milestone() {}
func (n *NodeUseAuthenticatorPassword) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*NodeUseAuthenticatorPassword) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputTakePassword{}, nil
}

func (i *NodeUseAuthenticatorPassword) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakePassword inputTakePassword
	if authflow.AsInput(input, &inputTakePassword) {
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

		return authflow.NewNodeSimple(&NodeDidVerifyAuthenticator{
			Authenticator:          info,
			PasswordChangeRequired: requireUpdate,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
