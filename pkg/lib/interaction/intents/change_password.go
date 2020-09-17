package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentChangePassword{})
}

type IntentChangePassword struct {
	UserID string                          `json:"user_id"`
	Stage  interaction.AuthenticationStage `json:"stage"`
}

func NewIntentChangePrimaryPassword(userID string) *IntentChangePassword {
	return &IntentChangePassword{
		UserID: userID,
		Stage:  interaction.AuthenticationStagePrimary,
	}
}

func NewIntentChangeSecondaryPassword(userID string) *IntentChangePassword {
	return &IntentChangePassword{
		UserID: userID,
		Stage:  interaction.AuthenticationStageSecondary,
	}
}

func (i *IntentChangePassword) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentChangePassword) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeChangePassword{
				Stage: i.Stage,
			},
		}, nil
	case *nodes.NodeChangePasswordEnd:
		// Password was not changed, ends the interaction
		return nil, nil
	case *nodes.NodeDoUpdateAuthenticator:
		return nil, nil
	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
