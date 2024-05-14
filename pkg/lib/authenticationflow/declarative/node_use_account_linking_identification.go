package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeUseAccountLinkingIdentification{})
}

type NodeUseAccountLinkingIdentification struct {
	Option   AccountLinkingIdentificationOption `json:"option,omitempty"`
	Conflict *AccountLinkingConflict            `json:"conflict,omitempty"`

	// oauth
	RedirectURI  string `json:"redirect_uri,omitempty"`
	ResponseMode string `json:"response_mode,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseAccountLinkingIdentification{}
var _ authflow.Milestone = &NodeUseAccountLinkingIdentification{}
var _ MilestoneUseAccountLinkingIdentification = &NodeUseAccountLinkingIdentification{}

func (*NodeUseAccountLinkingIdentification) Kind() string {
	return "NodeUseAccountLinkingIdentificationOption"
}

func (*NodeUseAccountLinkingIdentification) Milestone() {}
func (n *NodeUseAccountLinkingIdentification) MilestoneUseAccountLinkingIdentification() *AccountLinkingConflict {
	return n.Conflict
}
func (n *NodeUseAccountLinkingIdentification) MilestoneUseAccountLinkingIdentificationSelectedOption() AccountLinkingIdentificationOption {
	return n.Option
}
func (n *NodeUseAccountLinkingIdentification) MilestoneUseAccountLinkingIdentificationRedirectURI() string {
	return n.RedirectURI
}
func (n *NodeUseAccountLinkingIdentification) MilestoneUseAccountLinkingIdentificationResponseMode() string {
	return n.ResponseMode
}
