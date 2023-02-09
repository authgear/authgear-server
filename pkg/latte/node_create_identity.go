package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeCreateIdentity{})
}

type NodeCreateIdentity struct {
	IdentityInfo *identity.Info `json:"identity_info,omitempty"`
}

func (n *NodeCreateIdentity) Kind() string {
	return "latte.NodeCreateIdentity"
}

func (n *NodeCreateIdentity) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			err := deps.Identities.Create(n.IdentityInfo)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (*NodeCreateIdentity) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeCreateIdentity) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeCreateIdentity) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
