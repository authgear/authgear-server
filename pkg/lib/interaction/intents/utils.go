package intents

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func mustFindNodeSelectIdentity(graph *interaction.Graph) *nodes.NodeSelectIdentityEnd {
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

func mustFindNodeDoUpdateIdentity(graph *interaction.Graph) *nodes.NodeDoUpdateIdentity {
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
