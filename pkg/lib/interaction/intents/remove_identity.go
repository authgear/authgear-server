package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentRemoveIdentity{})
}

type IntentRemoveIdentity struct {
	UserID string `json:"user_id"`
}

func NewIntentRemoveIdentity(userID string) *IntentRemoveIdentity {
	return &IntentRemoveIdentity{
		UserID: userID,
	}
}

func (i *IntentRemoveIdentity) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentRemoveIdentity) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeRemoveIdentity{},
		}, nil

	case *nodes.NodeRemoveIdentity:
		return []interaction.Edge{
			&nodes.EdgeDoRemoveIdentity{Identity: node.IdentityInfo},
		}, nil
	case *nodes.NodeDoRemoveIdentity:
		return []interaction.Edge{
			&nodes.EdgeRemoveAuthenticator{IdentityInfo: node.Identity},
		}, nil

	case *nodes.NodeRemoveAuthenticator:
		return []interaction.Edge{
			&nodes.EdgeDoRemoveAuthenticator{Authenticators: node.Authenticators},
		}, nil
	case *nodes.NodeDoRemoveAuthenticator:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
