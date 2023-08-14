package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeUseRecoveryCode{})
}

type NodeUseRecoveryCode struct {
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

var _ workflow.NodeSimple = &NodeUseRecoveryCode{}
var _ workflow.Milestone = &NodeUseRecoveryCode{}
var _ MilestoneAuthenticationMethod = &NodeUseRecoveryCode{}
var _ workflow.InputReactor = &NodeUseRecoveryCode{}

func (*NodeUseRecoveryCode) Kind() string {
	return "workflowconfig.NodeUseRecoveryCode"
}

func (*NodeUseRecoveryCode) Milestone() {}
func (n *NodeUseRecoveryCode) MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return n.Authentication
}

func (*NodeUseRecoveryCode) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return []workflow.Input{&InputTakeRecoveryCode{}}, nil
}

func (n *NodeUseRecoveryCode) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeRecoveryCode inputTakeRecoveryCode
	if workflow.AsInput(input, &inputTakeRecoveryCode) {
		recoveryCode := inputTakeRecoveryCode.GetRecoveryCode()

		rc, err := deps.MFA.VerifyRecoveryCode(n.UserID, recoveryCode)
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(&NodeDoConsumeRecoveryCode{
			RecoveryCode: rc,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}
