package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoUseAuthenticator{})
}

type NodeDoUseAuthenticator struct {
	Authenticator          *authenticator.Info `json:"authenticator,omitempty"`
	PasswordChangeRequired bool                `json:"password_change_required,omitempty"`
}

var _ Milestone = &NodeDoUseAuthenticator{}

func (*NodeDoUseAuthenticator) Milestone() {}

var _ MilestoneDoUseAuthenticator = &NodeDoUseAuthenticator{}

func (n *NodeDoUseAuthenticator) MilestoneDoUseAuthenticator() (*NodeDoUseAuthenticator, bool) {
	return n, true
}

var _ MilestoneAuthenticated = &NodeDoUseAuthenticator{}

func (*NodeDoUseAuthenticator) MilestoneAuthenticated() {}

var _ workflow.NodeSimple = &NodeDoUseAuthenticator{}

func (*NodeDoUseAuthenticator) Kind() string {
	return "workflowconfig.NodeDoUseAuthenticator"
}

func (*NodeDoUseAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeDoUseAuthenticator) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoUseAuthenticator) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, inut workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoUseAuthenticator) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
