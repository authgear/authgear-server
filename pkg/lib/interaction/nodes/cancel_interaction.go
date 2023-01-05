package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCancelInteraction{})
}

type EdgeCancelInteraction struct {
}

func (e *EdgeCancelInteraction) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	if i, ok := graph.Intent.(interaction.IntentWithCancelURI); ok {
		return &NodeCancelInteraction{
			RedirectURI: i.GetCancelURI(),
		}, nil
	} else {
		// Not a cancelable intent, stuck it in previous node
		return nil, interaction.ErrIncompatibleInput
	}
}

type NodeCancelInteraction struct {
	RedirectURI string `json:"redirect_uri"`
}

// GetRedirectURI implements RedirectURIGetter.
func (n *NodeCancelInteraction) GetRedirectURI() string {
	return n.RedirectURI
}

func (n *NodeCancelInteraction) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCancelInteraction) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCancelInteraction) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{}, nil
}
