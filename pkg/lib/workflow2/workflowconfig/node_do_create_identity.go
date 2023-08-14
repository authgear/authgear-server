package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoCreateIdentity{})
}

type NodeDoCreateIdentity struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

var _ MilestoneDoCreateIdentity = &NodeDoCreateIdentity{}

func (*NodeDoCreateIdentity) Milestone() {}
func (n *NodeDoCreateIdentity) MilestoneDoCreateIdentity() *identity.Info {
	return n.Identity
}

var _ workflow.NodeSimple = &NodeDoCreateIdentity{}
var _ workflow.EffectGetter = &NodeDoCreateIdentity{}

func (n *NodeDoCreateIdentity) Kind() string {
	return "workflowconfig.NodeDoCreateIdentity"
}

func (n *NodeDoCreateIdentity) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			err := deps.Identities.Create(n.Identity)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}
