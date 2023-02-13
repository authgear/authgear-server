package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticateEmailLoginLink{})
}

type NodeAuthenticateEmailLoginLink struct {
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (n *NodeAuthenticateEmailLoginLink) Kind() string {
	return "latte.NodeAuthenticateEmailLoginLink"
}

func (n *NodeAuthenticateEmailLoginLink) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticateEmailLoginLink) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputCheckLoginLinkVerified{},
		&InputResendCode{},
	}, nil
}

func (n *NodeAuthenticateEmailLoginLink) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputCheckLoginLinkVerified inputCheckLoginLinkVerified
	var inputResendCode inputResendCode
	switch {
	case workflow.AsInput(input, &inputResendCode):
		info := n.Authenticator
		_, err := (&SendOOBCode{
			Deps:                 deps,
			Stage:                authenticatorKindToStage(info.Kind),
			IsAuthenticating:     true,
			AuthenticatorInfo:    info,
			IgnoreRatelimitError: false,
			OTPMode:              otp.OTPModeMagicLink,
		}).Do()
		if err != nil {
			return nil, err
		}
		return nil, workflow.ErrSameNode
	case workflow.AsInput(input, &inputCheckLoginLinkVerified):
		info := n.Authenticator
		_, err := deps.OTPCodes.VerifyMagicLinkCodeByTarget(info.OOBOTP.Email, true)
		if err != nil {
			if errors.Is(err, otp.ErrInvalidCode) {
				err = api.ErrInvalidCredentials
			}
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeVerifiedAuthenticator{
			Authenticator: info,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticateEmailLoginLink) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
