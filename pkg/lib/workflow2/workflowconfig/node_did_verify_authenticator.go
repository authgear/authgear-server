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

var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDidVerifyAuthenticator{}

func (n *NodeDidVerifyAuthenticator) MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodFromAuthenticatorType(n.Authenticator.Type)
}

var _ workflow.NodeSimple = &NodeDidVerifyAuthenticator{}

func (*NodeDidVerifyAuthenticator) Kind() string {
	return "workflowconfig.NodeDidVerifyAuthenticator"
}
