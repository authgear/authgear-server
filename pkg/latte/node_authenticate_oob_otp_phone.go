package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeAuthenticateOOBOTPPhone{})
}

type NodeAuthenticateOOBOTPPhone struct {
	Authenticator *authenticator.Info `json:"authenticator"`
}

func (n *NodeAuthenticateOOBOTPPhone) Kind() string {
	return "latte.NodeAuthenticateOOBOTPPhone"
}

func (n *NodeAuthenticateOOBOTPPhone) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeAuthenticateOOBOTPPhone) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return []workflow.Input{
		&InputTakeOOBOTPCode{},
		&InputResendCode{},
	}, nil
}

func (n *NodeAuthenticateOOBOTPPhone) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	var inputTakeOOBOTPCode inputTakeOOBOTPCode
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
		}).Do()
		if err != nil {
			return nil, err
		}
		return nil, workflow.ErrSameNode
	case workflow.AsInput(input, &inputTakeOOBOTPCode):
		info := n.Authenticator
		_, err := deps.Authenticators.VerifyWithSpec(info, &authenticator.Spec{
			OOBOTP: &authenticator.OOBOTPSpec{
				Code: inputTakeOOBOTPCode.GetCode(),
			},
		})
		if errors.Is(err, authenticator.ErrInvalidCredentials) {
			if err := DispatchAuthenticationFailedEvent(deps.Events, info); err != nil {
				return nil, err
			}
			return nil, api.ErrInvalidCredentials
		} else if err != nil {
			return nil, err
		}
		return workflow.NewNodeSimple(&NodeVerifiedAuthenticator{
			Authenticator: info,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeAuthenticateOOBOTPPhone) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return map[string]interface{}{}, nil
}
