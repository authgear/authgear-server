package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	workflow.RegisterNode(&NodeSendForgotPasswordCode{})
}

type NodeSendForgotPasswordCode struct {
	LoginID       string `json:"login_id"`
	OutputLoginID bool   `json:"output_login_id"`
}

func (n *NodeSendForgotPasswordCode) Kind() string {
	return "latte.NodeSendForgotPasswordCode"
}

func (n *NodeSendForgotPasswordCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*NodeSendForgotPasswordCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeSendForgotPasswordCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeSendForgotPasswordCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	output := map[string]interface{}{}

	if n.OutputLoginID {
		maskedLoginID := mail.MaskAddress(n.LoginID)
		if maskedLoginID == "" {
			maskedLoginID = phone.Mask(n.LoginID)
		}
		output["masked_login_id"] = maskedLoginID
	}

	return output, nil
}

func (n *NodeSendForgotPasswordCode) sendCode(ctx context.Context, deps *workflow.Dependencies) error {
	return deps.ForgotPassword.SendCode(n.LoginID)

}
