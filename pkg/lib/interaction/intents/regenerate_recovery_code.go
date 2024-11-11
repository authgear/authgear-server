package intents

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentRegenerateRecoveryCode{})
}

type IntentRegenerateRecoveryCode struct {
	UserID string `json:"user_id"`
}

func NewIntentRegenerateRecoveryCode(userID string) *IntentRegenerateRecoveryCode {
	return &IntentRegenerateRecoveryCode{
		UserID: userID,
	}
}

func (i *IntentRegenerateRecoveryCode) InstantiateRootNode(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(goCtx, ctx, graph, i)
}

func (i *IntentRegenerateRecoveryCode) DeriveEdgesForNode(goCtx context.Context, graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeGenerateRecoveryCode{IsRegenerate: true},
		}, nil

	case *nodes.NodeGenerateRecoveryCodeEnd:
		return []interaction.Edge{
			&nodes.EdgeDoGenerateRecoveryCode{
				RecoveryCodes: node.RecoveryCodes,
			},
		}, nil

	case *nodes.NodeDoGenerateRecoveryCode:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
