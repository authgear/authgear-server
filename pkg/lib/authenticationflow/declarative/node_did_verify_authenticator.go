package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDidVerifyAuthenticator{})
}

type NodeDidVerifyAuthenticator struct {
	Authenticator          *authenticator.Info `json:"authenticator,omitempty"`
	PasswordChangeRequired bool                `json:"password_change_required,omitempty"`
}

var _ authflow.NodeSimple = &NodeDidVerifyAuthenticator{}
var _ authflow.Milestone = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidVerifyAuthenticator = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidSelectAuthenticator = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidAuthenticate = &NodeDidVerifyAuthenticator{}
var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDidVerifyAuthenticator{}

func (*NodeDidVerifyAuthenticator) Kind() string {
	return "NodeDidVerifyAuthenticator"
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
