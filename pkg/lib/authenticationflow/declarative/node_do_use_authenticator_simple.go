package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoUseAuthenticatorSimple{})
}

type NodeDoUseAuthenticatorSimple struct {
	Authenticator *authenticator.Info `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseAuthenticatorSimple{}
var _ authflow.Milestone = &NodeDoUseAuthenticatorSimple{}
var _ MilestoneDidSelectAuthenticator = &NodeDoUseAuthenticatorSimple{}
var _ MilestoneDidAuthenticate = &NodeDoUseAuthenticatorSimple{}
var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDoUseAuthenticatorSimple{}

func (*NodeDoUseAuthenticatorSimple) Kind() string {
	return "NodeDoUseAuthenticatorSimple"
}

func (*NodeDoUseAuthenticatorSimple) Milestone() {}
func (n *NodeDoUseAuthenticatorSimple) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoUseAuthenticatorSimple) MilestoneDidAuthenticate() (amr []string) {
	return n.Authenticator.AMR()
}
func (n *NodeDoUseAuthenticatorSimple) MilestoneDidAuthenticateAuthenticator() (*authenticator.Info, bool) {
	return n.Authenticator, true
}
func (n *NodeDoUseAuthenticatorSimple) MilestoneDidAuthenticateAuthentication() (*model.Authentication, bool) {
	authn := n.Authenticator.ToAuthentication()
	authnModel := n.Authenticator.ToModel()
	return &model.Authentication{
		Authentication: authn,
		Authenticator:  &authnModel,
	}, true
}
func (n *NodeDoUseAuthenticatorSimple) MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodFromAuthenticatorType(n.Authenticator.Type)
}
