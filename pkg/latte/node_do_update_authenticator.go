package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoUpdateAuthenticator{})
}

type NodeDoUpdateAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (n *NodeDoUpdateAuthenticator) Kind() string {
	return "latte.NodeDoUpdateAuthenticator"
}

func (n *NodeDoUpdateAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.Authenticators.Update(ctx, n.Authenticator)
		}),
	}, nil
}

func (*NodeDoUpdateAuthenticator) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoUpdateAuthenticator) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoUpdateAuthenticator) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
