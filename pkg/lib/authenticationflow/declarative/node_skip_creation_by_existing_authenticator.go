package declarative

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeSkipCreationByExistingAuthenticator{})
}

type NodeSkipCreationByExistingAuthenticator struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Authenticator  *authenticator.Info                     `json:"authenticator,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeSkipCreationByExistingAuthenticator{}
var _ authflow.Milestone = &NodeSkipCreationByExistingAuthenticator{}
var _ MilestoneAuthenticationMethod = &NodeSkipCreationByExistingAuthenticator{}
var _ MilestoneDoCreateAuthenticator = &NodeSkipCreationByExistingAuthenticator{}
var _ MilestoneFlowCreateAuthenticator = &NodeSkipCreationByExistingAuthenticator{}

func (*NodeSkipCreationByExistingAuthenticator) Kind() string {
	return "NodeSkipCreationByExistingAuthenticator"
}

func (n *NodeSkipCreationByExistingAuthenticator) Milestone() {}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneFlowCreateAuthenticator(flows authflow.Flows) (MilestoneDoCreateAuthenticator, authflow.Flows, bool) {
	return n, flows, true
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDoCreateAuthenticator() *authenticator.Info {
	return n.Authenticator
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDoCreateAuthenticatorSkipCreate() {
	// Already skipping
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDoCreateAuthenticatorUpdate(newInfo *authenticator.Info) {
	n.Authenticator = newInfo
}
