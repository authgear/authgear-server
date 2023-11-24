package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeUseAccountRecoveryDestination{})
}

type NodeUseAccountRecoveryDestination struct {
	Destination *AccountRecoveryDestinationOptionInternal `json:"destination,omitempty"`
}

var _ authflow.Milestone = &NodeUseAccountRecoveryDestination{}
var _ MilestoneDoUseAccountRecoveryDestination = &NodeUseAccountRecoveryDestination{}

var _ authflow.NodeSimple = &NodeUseAccountRecoveryDestination{}

func (*NodeUseAccountRecoveryDestination) Kind() string {
	return "NodeUseAccountRecoveryDestination"
}

func (*NodeUseAccountRecoveryDestination) Milestone() {}
func (n *NodeUseAccountRecoveryDestination) MilestoneDoUseAccountRecoveryDestination() *AccountRecoveryDestinationOptionInternal {
	return n.Destination
}
