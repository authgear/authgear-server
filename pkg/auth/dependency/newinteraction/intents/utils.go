package intents

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func mustFindNodeSelectIdentity(graph *newinteraction.Graph) *nodes.NodeSelectIdentityEnd {
	var selectIdentity *nodes.NodeSelectIdentityEnd
	for _, node := range graph.Nodes {
		if node, ok := node.(*nodes.NodeSelectIdentityEnd); ok {
			selectIdentity = node
			break
		}
	}
	if selectIdentity == nil {
		panic("interaction: expect identity already selected")
	}
	return selectIdentity
}

func mustFindNodeDoUpdateIdentity(graph *newinteraction.Graph) *nodes.NodeDoUpdateIdentity {
	var doUpdateIdentity *nodes.NodeDoUpdateIdentity
	for _, node := range graph.Nodes {
		if node, ok := node.(*nodes.NodeDoUpdateIdentity); ok {
			doUpdateIdentity = node
			break
		}
	}
	if doUpdateIdentity == nil {
		panic("interaction: expect identity already updated")
	}
	return doUpdateIdentity
}
