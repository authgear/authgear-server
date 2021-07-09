package intents

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func mustFindNodeSelectIdentity(graph *interaction.Graph) *nodes.NodeSelectIdentityEnd {
	var selectIdentity *nodes.NodeSelectIdentityEnd
	for _, annotatedNode := range graph.AnnotatedNodes {
		if node, ok := annotatedNode.Node.(*nodes.NodeSelectIdentityEnd); ok {
			selectIdentity = node
			break
		}
	}
	if selectIdentity == nil {
		panic("interaction: expect identity already selected")
	}
	return selectIdentity
}
