package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoCreateAuthenticator{})
}

type NodeDoCreateAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

func (n *NodeDoCreateAuthenticator) Kind() string {
	return "latte.NodeDoCreateAuthenticator"
}

func (n *NodeDoCreateAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.Authenticators.Create(ctx, n.Authenticator, false)
		}),
	}, nil
}

func (*NodeDoCreateAuthenticator) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoCreateAuthenticator) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoCreateAuthenticator) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
