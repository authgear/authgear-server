package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDidPerformBotProtectionVerification{})
}

type NodeDidPerformBotProtectionVerification struct {
}

var _ authflow.NodeSimple = &NodeDidPerformBotProtectionVerification{}
var _ MilestoneDidPerformBotProtectionVerification = &NodeDidPerformBotProtectionVerification{}

func (n *NodeDidPerformBotProtectionVerification) Kind() string {
	return "NodeDidPerformBotProtectionVerification"
}

func (*NodeDidPerformBotProtectionVerification) Milestone() {}
func (n *NodeDidPerformBotProtectionVerification) MilestoneDidPerformBotProtectionVerification() {
}

func NewNodeDidPerformBotProtectionVerification(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDidPerformBotProtectionVerification) (*NodeDidPerformBotProtectionVerification, error) {
	return n, nil
}
