package workflowconfig

import (
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeSentinel{})
}

// NodeSentinel is useful to terminate a workflow.
type NodeSentinel struct{}

var _ workflow.NodeSimple = &NodeSentinel{}

func (n *NodeSentinel) Kind() string {
	return "workflowconfig.NodeSentinel"
}
