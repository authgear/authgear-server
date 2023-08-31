package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeSentinel{})
}

// NodeSentinel is useful to terminate a flow.
type NodeSentinel struct{}

var _ authflow.NodeSimple = &NodeSentinel{}

func (n *NodeSentinel) Kind() string {
	return "NodeSentinel"
}
