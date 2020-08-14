package nodes

import "github.com/authgear/authgear-server/pkg/lib/interaction"

func getIdentityConflictNode(graph *interaction.Graph) (*NodeCheckIdentityConflict, bool) {
	for _, node := range graph.Nodes {
		if node, ok := node.(*NodeCheckIdentityConflict); ok {
			return node, true
		}
	}
	return nil, false
}
