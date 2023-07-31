package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoUseIdentity{})
}

type NodeDoUseIdentity struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

func (*NodeDoUseIdentity) Kind() string {
	return "workflowconfig.NodeDoUseIdentity"
}

func (*NodeDoUseIdentity) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeDoUseIdentity) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoUseIdentity) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, inut workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoUseIdentity) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
