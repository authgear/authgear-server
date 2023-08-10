package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoUseDeviceToken{})
}

type NodeDoUseDeviceToken struct{}

var _ Milestone = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) Milestone() {}

var _ MilestoneAuthenticationMethod = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) MilestoneAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return config.WorkflowAuthenticationMethodDeviceToken
}

var _ MilestoneDidAuthenticate = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) MilestoneDidAuthenticate() {}

var _ workflow.NodeSimple = &NodeDoUseDeviceToken{}

func (*NodeDoUseDeviceToken) Kind() string {
	return "workflowconfig.NodeDoUseDeviceToken"
}

func (*NodeDoUseDeviceToken) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeDoUseDeviceToken) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoUseDeviceToken) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoUseDeviceToken) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
