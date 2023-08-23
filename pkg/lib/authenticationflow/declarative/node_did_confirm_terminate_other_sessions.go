package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

func init() {
	authflow.RegisterNode(&NodeDidConfirmTerminateOtherSessions{})
}

type NodeDidConfirmTerminateOtherSessions struct{}

var _ authflow.NodeSimple = &NodeDidConfirmTerminateOtherSessions{}

func (n *NodeDidConfirmTerminateOtherSessions) Kind() string {
	return "NodeDidConfirmTerminateOtherSessions"
}
