package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeDidSelectAuthenticator{})
}

type NodeDidSelectAuthenticator struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeDidSelectAuthenticator{}
var _ authflow.Milestone = &NodeDidSelectAuthenticator{}
var _ MilestoneDidSelectAuthenticator = &NodeDidSelectAuthenticator{}
var _ MilestoneSwitchToExistingUser = &NodeDidSelectAuthenticator{}

func (*NodeDidSelectAuthenticator) Kind() string {
	return "NodeDidSelectAuthenticator"
}

func (n *NodeDidSelectAuthenticator) Milestone() {}
func (n *NodeDidSelectAuthenticator) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (i *NodeDidSelectAuthenticator) MilestoneSwitchToExistingUser(newUserID string) {
	i.Authenticator = i.Authenticator.UpdateUserID(newUserID)
}
