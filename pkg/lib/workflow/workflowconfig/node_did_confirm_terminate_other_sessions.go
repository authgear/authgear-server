package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDidConfirmTerminateOtherSessions{})
}

type NodeDidConfirmTerminateOtherSessions struct{}

var _ workflow.NodeSimple = &NodeDidConfirmTerminateOtherSessions{}

func (n *NodeDidConfirmTerminateOtherSessions) Kind() string {
	return "workflowconfig.NodeDidConfirmTerminateOtherSessions"
}

func (n *NodeDidConfirmTerminateOtherSessions) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeDidConfirmTerminateOtherSessions) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDidConfirmTerminateOtherSessions) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDidConfirmTerminateOtherSessions) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
