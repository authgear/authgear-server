package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoCreateAuthenticator{})
}

type NodeDoCreateAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ MilestoneDoCreateAuthenticator = &NodeDoCreateAuthenticator{}

func (n *NodeDoCreateAuthenticator) Milestone() {}
func (n *NodeDoCreateAuthenticator) MilestoneDoCreateAuthenticator() *authenticator.Info {
	return n.Authenticator
}

var _ workflow.NodeSimple = &NodeDoCreateAuthenticator{}
var _ workflow.EffectGetter = &NodeDoCreateAuthenticator{}

func (n *NodeDoCreateAuthenticator) Kind() string {
	return "workflowconfig.NodeDoCreateAuthenticator"
}

func (n *NodeDoCreateAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.Authenticators.Create(n.Authenticator, false)
		}),
	}, nil
}
