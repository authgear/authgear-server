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
	authflow.RegisterNode(&NodeUseAuthenticatorTOTP{})
}

type NodeUseAuthenticatorTOTP struct {
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAuthenticatorTOTP{}
var _ authflow.Milestone = &NodeUseAuthenticatorTOTP{}
var _ MilestoneAuthenticationMethod = &NodeUseAuthenticatorTOTP{}
var _ authflow.InputReactor = &NodeUseAuthenticatorTOTP{}

func (*NodeUseAuthenticatorTOTP) Kind() string {
	return "NodeUseAuthenticatorTOTP"
}

func (*NodeUseAuthenticatorTOTP) Milestone() {}
func (n *NodeUseAuthenticatorTOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*NodeUseAuthenticatorTOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputTakeTOTP{}, nil
}

func (n *NodeUseAuthenticatorTOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeTOTP inputTakeTOTP
	if authflow.AsInput(input, &inputTakeTOTP) {
		as, err := deps.Authenticators.List(
			n.UserID,
			authenticator.KeepKind(n.Authentication.AuthenticatorKind()),
			authenticator.KeepType(model.AuthenticatorTypeTOTP),
		)
		if err != nil {
			return nil, err
		}

		code := inputTakeTOTP.GetCode()
		spec := &authenticator.Spec{
			TOTP: &authenticator.TOTPSpec{
				Code: code,
			},
		}

		info, _, err := deps.Authenticators.VerifyOneWithSpec(
			as,
			spec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					n.UserID,
					authn.AuthenticationStageFromAuthenticationMethod(n.Authentication),
					authn.AuthenticationTypeTOTP,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDidVerifyAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
