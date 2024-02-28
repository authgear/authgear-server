package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
)

func init() {
	interaction.RegisterNode(&NodeSettingsActionEnd{})
}

type EdgeSettingsActionEnd struct{}

func (e *EdgeSettingsActionEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeSettingsActionEnd{}, nil
}

type NodeSettingsActionEnd struct{}

func (n *NodeSettingsActionEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeSettingsActionEnd) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			entry, err := ctx.OAuthSessions.Get(ctx.OAuthSessionID)
			if errors.Is(err, oauthsession.ErrNotFound) {
				return nil
			} else if err != nil {
				return err
			}

			entry.T.SettingsActionResult = oauthsession.NewSettingsActionResult()
			err = ctx.OAuthSessions.Save(entry)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeSettingsActionEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
