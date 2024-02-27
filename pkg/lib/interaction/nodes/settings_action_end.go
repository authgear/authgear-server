package nodes

import (
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
			if err != nil {
				return err
			}

			if entry.T.AuthorizationRequest.ResponseType() != "settings_action" {
				return nil
			}

			entry.T.SettingsActionResult = oauthsession.NewSettingsActionResult()
			ctx.OAuthSessions.Save(entry)

			return nil
		}),
	}, nil
}

func (n *NodeSettingsActionEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
