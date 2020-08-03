package nodes

import "github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"

func getIdentityConflictNode(graph *newinteraction.Graph) (*NodeCheckIdentityConflict, bool) {
	for _, node := range graph.Nodes {
		if node, ok := node.(*NodeCheckIdentityConflict); ok {
			return node, true
		}
	}
	return nil, false
}
