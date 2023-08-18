package workflowconfig

import (
	"context"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeGenerateRecoveryCode{})
}

type NodeGenerateRecoveryCodeData struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

var _ workflow.Data = &NodeGenerateRecoveryCodeData{}

func (m NodeGenerateRecoveryCodeData) Data() {}

type NodeGenerateRecoveryCode struct {
	UserID        string   `json:"user_id,omitempty"`
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

var _ workflow.NodeSimple = &NodeGenerateRecoveryCode{}
var _ workflow.InputReactor = &NodeGenerateRecoveryCode{}
var _ workflow.DataOutputer = &NodeGenerateRecoveryCode{}

func NewNodeGenerateRecoveryCode(deps *workflow.Dependencies, n *NodeGenerateRecoveryCode) *NodeGenerateRecoveryCode {
	n.RecoveryCodes = deps.MFA.GenerateRecoveryCodes()
	return n
}

func (*NodeGenerateRecoveryCode) Kind() string {
	return "workflowconfig.NodeGenerateRecoveryCode"
}

func (*NodeGenerateRecoveryCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	return &InputConfirmRecoveryCode{}, nil
}

func (n *NodeGenerateRecoveryCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputConfirmRecoveryCode inputConfirmRecoveryCode
	if workflow.AsInput(input, &inputConfirmRecoveryCode) {
		return workflow.NewNodeSimple(&NodeDoReplaceRecoveryCode{
			UserID:        n.UserID,
			RecoveryCodes: n.RecoveryCodes,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeGenerateRecoveryCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.Data, error) {
	return NodeGenerateRecoveryCodeData{
		RecoveryCodes: n.RecoveryCodes,
	}, nil
}
