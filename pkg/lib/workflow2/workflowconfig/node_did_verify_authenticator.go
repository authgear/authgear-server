package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDidVerifyAuthenticator{})
}

type NodeDidVerifyAuthenticator struct {
	Authenticator          *authenticator.Info `json:"authenticator,omitempty"`
	PasswordChangeRequired bool                `json:"password_change_required,omitempty"`
}

var _ workflow.NodeSimple = &NodeDidVerifyAuthenticator{}
var _ workflow.Milestone = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidVerifyAuthenticator = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidSelectAuthenticator = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidAuthenticate = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDidVerifyAuthenticator{}

func (*NodeDidVerifyAuthenticator) Kind() string {
	return "workflowconfig.NodeDidVerifyAuthenticator"
}

func (*NodeDidVerifyAuthenticator) Milestone() {}
func (n *NodeDidVerifyAuthenticator) MilestoneDidVerifyAuthenticator() *NodeDidVerifyAuthenticator {
	return n
}
func (n *NodeDidVerifyAuthenticator) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDidVerifyAuthenticator) MilestoneDidAuthenticate() (amr []string) {
	return n.Authenticator.AMR()
}
func (n *NodeDidVerifyAuthenticator) MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodFromAuthenticatorType(n.Authenticator.Type)
}
