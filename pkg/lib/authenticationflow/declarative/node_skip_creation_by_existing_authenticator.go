package declarative

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func init() {
	authflow.RegisterNode(&NodeSkipCreationByExistingAuthenticator{})
}

type NodeSkipCreationByExistingAuthenticator struct {
	JSONPointer    jsonpointer.T                          `json:"json_pointer,omitempty"`
	Authenticator  *authenticator.Info                    `json:"authenticator,omitempty"`
	Authentication model.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.NodeSimple = &NodeSkipCreationByExistingAuthenticator{}
var _ authflow.Milestone = &NodeSkipCreationByExistingAuthenticator{}
var _ MilestoneFlowSelectAuthenticationMethod = &NodeSkipCreationByExistingAuthenticator{}
var _ MilestoneDidSelectAuthenticationMethod = &NodeSkipCreationByExistingAuthenticator{}
var _ MilestoneDoCreateAuthenticator = &NodeSkipCreationByExistingAuthenticator{}
var _ MilestoneFlowCreateAuthenticator = &NodeSkipCreationByExistingAuthenticator{}

func (*NodeSkipCreationByExistingAuthenticator) Kind() string {
	return "NodeSkipCreationByExistingAuthenticator"
}

func (n *NodeSkipCreationByExistingAuthenticator) Milestone() {}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneFlowCreateAuthenticator(flows authflow.Flows) (MilestoneDoCreateAuthenticator, authflow.Flows, bool) {
	return n, flows, true
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (MilestoneDidSelectAuthenticationMethod, authflow.Flows, bool) {
	return n, flows, true
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDidSelectAuthenticationMethod() model.AuthenticationFlowAuthentication {
	return n.Authentication
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDoCreateAuthenticator() (*authenticator.Info, bool) {
	return n.Authenticator, false
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDoCreateAuthenticatorSkipCreate() {
	// Already skipping
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDoCreateAuthenticatorAuthentication() (*model.Authentication, bool) {
	authnModel := n.Authenticator.ToModel()
	return &model.Authentication{
		Authentication: n.Authentication,
		Authenticator:  &authnModel,
	}, true
}
func (n *NodeSkipCreationByExistingAuthenticator) MilestoneDoCreateAuthenticatorUpdate(newInfo *authenticator.Info) {
	n.Authenticator = newInfo
}
