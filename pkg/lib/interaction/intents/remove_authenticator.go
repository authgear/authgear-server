package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentRemoveAuthenticator{})
}

type IntentRemoveAuthenticator struct {
	UserID string `json:"user_id"`
}

func NewIntentRemoveAuthenticator(userID string) *IntentRemoveAuthenticator {
	return &IntentRemoveAuthenticator{
		UserID: userID,
	}
}

func (i *IntentRemoveAuthenticator) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentRemoveAuthenticator) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeRemoveAuthenticator{},
		}, nil

	case *nodes.NodeRemoveAuthenticator:
		return []interaction.Edge{
			&nodes.EdgeDoRemoveAuthenticator{
				Authenticator:        node.AuthenticatorInfo,
				BypassMFARequirement: node.BypassMFARequirement,
			},
		}, nil

	case *nodes.NodeDoRemoveAuthenticator:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
