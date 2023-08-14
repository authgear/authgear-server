package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoUpdateAuthenticator{})
}

type NodeDoUpdateAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ workflow.NodeSimple = &NodeDoUpdateAuthenticator{}
var _ workflow.EffectGetter = &NodeDoUpdateAuthenticator{}

func (*NodeDoUpdateAuthenticator) Kind() string {
	return "workflowconfig.NodeDoUpdateAuthenticator"
}

func (n *NodeDoUpdateAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.Authenticators.Update(n.Authenticator)
		}),
	}, nil
}
