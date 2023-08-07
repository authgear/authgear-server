package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeUseRecoveryCode{})
}

type NodeUseRecoveryCode struct {
	UserID         string                              `json:"user_id,omitempty"`
	Authentication config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
}

var _ MilestoneAuthenticationMethod = &NodeUseRecoveryCode{}

func (*NodeUseRecoveryCode) Milestone() {}
func (n *NodeUseRecoveryCode) MilestoneAuthenticationMethod() (config.WorkflowAuthenticationMethod, bool) {
	return n.Authentication, true
}

var _ workflow.NodeSimple = &NodeUseRecoveryCode{}

func (*NodeUseRecoveryCode) Kind() string {
	return "workflowconfig.NodeUseRecoveryCode"
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

func (*NodeUseRecoveryCode) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeUseRecoveryCode) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
