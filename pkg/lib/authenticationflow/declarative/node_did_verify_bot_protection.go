package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDidVerifyBotProtection{})
}

type NodeDidVerifyBotProtection struct {
}

var _ authflow.NodeSimple = &NodeDidVerifyBotProtection{}
var _ MilestoneDidVerifyBotProtection = &NodeDidVerifyBotProtection{}

func (n *NodeDidVerifyBotProtection) Kind() string {
	return "NodeDidVerifyBotProtection"
}

func (*NodeDidVerifyBotProtection) Milestone() {}
func (n *NodeDidVerifyBotProtection) MilestoneDidVerifyBotProtection() {
}

func NewNodeDidVerifyBotProtection(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDidVerifyBotProtection) (*NodeDidVerifyBotProtection, error) {
	return n, nil
}
