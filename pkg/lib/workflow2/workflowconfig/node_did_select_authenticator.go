package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDidSelectAuthenticator{})
}

type NodeDidSelectAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ workflow.NodeSimple = &NodeDidSelectAuthenticator{}
var _ workflow.Milestone = &NodeDidSelectAuthenticator{}
var _ MilestoneDidSelectAuthenticator = &NodeDidSelectAuthenticator{}

func (*NodeDidSelectAuthenticator) Kind() string {
	return "workflowconfig.NodeDidSelectAuthenticator"
}

func (n *NodeDidSelectAuthenticator) Milestone() {}
func (n *NodeDidSelectAuthenticator) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
