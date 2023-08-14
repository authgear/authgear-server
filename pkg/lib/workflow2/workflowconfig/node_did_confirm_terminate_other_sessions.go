package workflowconfig

import (
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDidConfirmTerminateOtherSessions{})
}

type NodeDidConfirmTerminateOtherSessions struct{}

var _ workflow.NodeSimple = &NodeDidConfirmTerminateOtherSessions{}

func (n *NodeDidConfirmTerminateOtherSessions) Kind() string {
	return "workflowconfig.NodeDidConfirmTerminateOtherSessions"
}
