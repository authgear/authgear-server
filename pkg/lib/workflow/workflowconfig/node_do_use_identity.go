package workflowconfig

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoUseIdentity{})
}

type UserIDGetter interface {
	GetUserID() string
}

type NodeDoUseIdentity struct {
	Identity *identity.Info `json:"identity,omitempty"`
}

var _ UserIDGetter = &NodeDoUseIdentity{}

func (n *NodeDoUseIdentity) GetUserID() string {
	return n.Identity.UserID
}

func NewNodeDoUseIdentity(workflows workflow.Workflows, n *NodeDoUseIdentity) (*NodeDoUseIdentity, error) {
	userID, err := getUserID(workflows)
	if errors.Is(err, ErrNoUserID) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	if userID != "" && userID != n.Identity.UserID {
		return nil, ErrDifferentUserID
	}

	return n, nil
}

func (*NodeDoUseIdentity) Kind() string {
	return "workflowconfig.NodeDoUseIdentity"
}

func (*NodeDoUseIdentity) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (*NodeDoUseIdentity) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoUseIdentity) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, inut workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (*NodeDoUseIdentity) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
