package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeSentinel{})
}

// NodeSentinel is useful to terminate a workflow.
type NodeSentinel struct{}

func (n *NodeSentinel) Kind() string {
	return "workflowconfig.NodeSentinel"
}

func (n *NodeSentinel) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeSentinel) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (n *NodeSentinel) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeSentinel) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
