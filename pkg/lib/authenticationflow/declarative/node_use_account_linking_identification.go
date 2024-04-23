package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

func init() {
	authflow.RegisterNode(&NodeUseAccountLinkingIdentification{})
}

type NodeUseAccountLinkingIdentification struct {
	Option   AccountLinkingIdentificationOption `json:"option,omitempty"`
	Identity *identity.Info                     `json:"identity,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAccountLinkingIdentification{}
var _ authflow.Milestone = &NodeUseAccountLinkingIdentification{}
var _ MilestoneUseAccountLinkingIdentification = &NodeUseAccountLinkingIdentification{}

func (*NodeUseAccountLinkingIdentification) Kind() string {
	return "NodeUseAccountLinkingIdentificationOption"
}

func (*NodeUseAccountLinkingIdentification) Milestone() {}
func (n *NodeUseAccountLinkingIdentification) MilestoneUseAccountLinkingIdentification() *identity.Info {
	return n.Identity
}
func (n *NodeUseAccountLinkingIdentification) MilestoneUseAccountLinkingIdentificationSelectedOption() AccountLinkingIdentificationOption {
	return n.Option
}
