package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentRemoveIdentity{})
}

type IntentRemoveIdentity struct {
	UserID string `json:"user_id"`
}

func NewIntentRemoveIdentity(userID string) *IntentRemoveIdentity {
	return &IntentRemoveIdentity{
		UserID: userID,
	}
}

func (i *IntentRemoveIdentity) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentRemoveIdentity) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []newinteraction.Edge{
			&nodes.EdgeRemoveIdentity{},
		}, nil

	case *nodes.NodeRemoveIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeDoRemoveIdentity{Identity: node.IdentityInfo},
		}, nil
	case *nodes.NodeDoRemoveIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeRemoveAuthenticator{IdentityInfo: node.Identity},
		}, nil

	case *nodes.NodeRemoveAuthenticator:
		return []newinteraction.Edge{
			&nodes.EdgeDoRemoveAuthenticator{Authenticators: node.Authenticators},
		}, nil
	case *nodes.NodeDoRemoveAuthenticator:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
