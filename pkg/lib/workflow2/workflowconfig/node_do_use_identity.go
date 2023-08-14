package workflowconfig

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoUseIdentity{})
}

type NodeDoUseIdentity struct {
	Identity *identity.Info `json:"identity,omitempty"`
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

var _ workflow.NodeSimple = &NodeDoUseIdentity{}
var _ workflow.Milestone = &NodeDoUseIdentity{}
var _ MilestoneDoUseUser = &NodeDoUseIdentity{}
var _ MilestoneDoUseIdentity = &NodeDoUseIdentity{}

func (*NodeDoUseIdentity) Kind() string {
	return "workflowconfig.NodeDoUseIdentity"
}

func (*NodeDoUseIdentity) Milestone() {}
func (n *NodeDoUseIdentity) MilestoneDoUseUser() string {
	return n.Identity.UserID
}
func (n *NodeDoUseIdentity) MilestoneDoUseIdentity() *identity.Info { return n.Identity }
