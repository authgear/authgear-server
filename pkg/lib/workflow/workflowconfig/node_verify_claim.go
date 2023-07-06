package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeVerifyClaim{})
}

type NodeVerifyClaim struct {
	UserID     string          `json:"user_id,omitempty"`
	ClaimName  model.ClaimName `json:"claim_name,omitempty"`
	ClaimValue string          `json:"claim_value,omitempty"`
}

func (n *NodeVerifyClaim) Kind() string {
	return "workflowconfig.NodeVerifyClaim"
}

func (n *NodeVerifyClaim) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (n *NodeVerifyClaim) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	// FIXME(workflow): resend code and verify incoming code.
	return nil, nil
}

func (n *NodeVerifyClaim) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	// FIXME(workflow): resend code and verify incoming code.
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeVerifyClaim) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	// FIXME(workflow): output state about verification.
	return nil, nil
}
