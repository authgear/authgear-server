package nodes

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
)

func init() {
	interaction.RegisterNode(&NodeSettingsActionEnd{})
}

type EdgeSettingsActionEnd struct{}

func (e *EdgeSettingsActionEnd) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	return &NodeSettingsActionEnd{}, nil
}

type NodeSettingsActionEnd struct{}

func (n *NodeSettingsActionEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeSettingsActionEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			entry, err := ctx.OAuthSessions.Get(goCtx, ctx.OAuthSessionID)
			if errors.Is(err, oauthsession.ErrNotFound) {
				return nil
			} else if err != nil {
				return err
			}

			entry.T.SettingsActionResult = oauthsession.NewSettingsActionResult()
			err = ctx.OAuthSessions.Save(goCtx, entry)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeSettingsActionEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}
