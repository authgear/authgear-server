package intents

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentChangePassword{})
}

type IntentChangePassword struct {
	UserID string                    `json:"user_id"`
	Stage  authn.AuthenticationStage `json:"stage"`
}

func NewIntentChangePrimaryPassword(userID string) *IntentChangePassword {
	return &IntentChangePassword{
		UserID: userID,
		Stage:  authn.AuthenticationStagePrimary,
	}
}

func NewIntentChangeSecondaryPassword(userID string) *IntentChangePassword {
	return &IntentChangePassword{
		UserID: userID,
		Stage:  authn.AuthenticationStageSecondary,
	}
}

func (i *IntentChangePassword) InstantiateRootNode(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(goCtx, ctx, graph, i)
}

func (i *IntentChangePassword) DeriveEdgesForNode(goCtx context.Context, graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeChangePassword{
				Stage: i.Stage,
			},
		}, nil
	case *nodes.NodeChangePasswordEnd:
		// We rely on NodeDoEnsureSession to write authentication info.
		return []interaction.Edge{
			&nodes.EdgeDoEnsureSession{
				Mode: nodes.EnsureSessionModeNoop,
			},
		}, nil
	case *nodes.NodeDoUpdateAuthenticator:
		// We rely on NodeDoEnsureSession to write authentication info.
		return []interaction.Edge{
			&nodes.EdgeDoEnsureSession{
				Mode: nodes.EnsureSessionModeNoop,
			},
		}, nil
	case *nodes.NodeDoEnsureSession:
		return []interaction.Edge{
			&nodes.EdgeSettingsActionEnd{},
		}, nil
	case *nodes.NodeSettingsActionEnd:
		// Intent is finished.
		return nil, nil
	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
