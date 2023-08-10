package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDidVerifyAuthenticator{})
}

type NodeDidVerifyAuthenticator struct {
	Authenticator          *authenticator.Info `json:"authenticator,omitempty"`
	PasswordChangeRequired bool                `json:"password_change_required,omitempty"`
}

var _ Milestone = &NodeDidVerifyAuthenticator{}

func (*NodeDidVerifyAuthenticator) Milestone() {}

var _ MilestoneDidVerifyAuthenticator = &NodeDidVerifyAuthenticator{}

func (n *NodeDidVerifyAuthenticator) MilestoneDidVerifyAuthenticator() *NodeDidVerifyAuthenticator {
	return n
}

var _ MilestoneDidSelectAuthenticator = &NodeDidVerifyAuthenticator{}

func (n *NodeDidVerifyAuthenticator) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}

var _ MilestoneDidAuthenticate = &NodeDidVerifyAuthenticator{}

func (n *NodeDidVerifyAuthenticator) MilestoneDidAuthenticate() (amr []string) {
	return n.Authenticator.AMR()
}

var _ workflow.NodeSimple = &NodeDidVerifyAuthenticator{}

func (*NodeDidVerifyAuthenticator) Kind() string {
	return "workflowconfig.NodeDidVerifyAuthenticator"
}

func (*NodeDidVerifyAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeDidVerifyAuthenticator) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDidVerifyAuthenticator) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, inut workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDidVerifyAuthenticator) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
