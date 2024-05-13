package declarative

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoUseAuthenticatorPassword{})
}

type NodeDoUseAuthenticatorPassword struct {
	JSONPointer            jsonpointer.T        `json:"json_pointer,omitempty"`
	Authenticator          *authenticator.Info  `json:"authenticator,omitempty"`
	PasswordChangeRequired bool                 `json:"password_change_required,omitempty"`
	PasswordChangeReason   PasswordChangeReason `json:"password_change_required_reason,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseAuthenticatorPassword{}
var _ authflow.Milestone = &NodeDoUseAuthenticatorPassword{}
var _ MilestoneDoUseAuthenticatorPassword = &NodeDoUseAuthenticatorPassword{}
var _ MilestoneDidSelectAuthenticator = &NodeDoUseAuthenticatorPassword{}
var _ MilestoneDidAuthenticate = &NodeDoUseAuthenticatorPassword{}
var _ MilestoneDidUseAuthenticationLockoutMethod = &NodeDoUseAuthenticatorPassword{}

func (*NodeDoUseAuthenticatorPassword) Kind() string {
	return "NodeDoUseAuthenticatorPassword"
}

func (*NodeDoUseAuthenticatorPassword) Milestone() {}
func (n *NodeDoUseAuthenticatorPassword) MilestoneDoUseAuthenticatorPassword() *NodeDoUseAuthenticatorPassword {
	return n
}
func (n *NodeDoUseAuthenticatorPassword) MilestoneDidSelectAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeDoUseAuthenticatorPassword) MilestoneDidAuthenticate() (amr []string) {
	return n.Authenticator.AMR()
}
func (n *NodeDoUseAuthenticatorPassword) MilestoneDidUseAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool) {
	return config.AuthenticationLockoutMethodFromAuthenticatorType(n.Authenticator.Type)
}

func (n *NodeDoUseAuthenticatorPassword) GetJSONPointer() jsonpointer.T {
	return n.JSONPointer
}
