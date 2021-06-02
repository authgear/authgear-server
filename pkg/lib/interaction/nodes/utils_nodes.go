package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func getIdentityConflictNode(graph *interaction.Graph) (*NodeCheckIdentityConflict, bool) {
	for _, node := range graph.Nodes {
		if node, ok := node.(*NodeCheckIdentityConflict); ok {
			return node, true
		}
	}
	return nil, false
}

// EdgeTerminal is used to indicate a terminal state of interaction; the
// interaction cannot further, and must be rewound to a previous step to
// continue.
type EdgeTerminal struct{}

func (e *EdgeTerminal) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	// Use ErrIncompatibleInput to 'stuck' the interaction at the current node.
	return nil, interaction.ErrIncompatibleInput
}

type InputAuthenticationStage interface {
	GetAuthenticationStage() authn.AuthenticationStage
}
