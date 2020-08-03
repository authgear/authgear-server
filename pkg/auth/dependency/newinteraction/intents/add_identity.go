package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentAddIdentity{})
}

type IntentAddIdentity struct {
	UserID string `json:"user_id"`
}

func NewIntentAddIdentity(userID string) *IntentAddIdentity {
	return &IntentAddIdentity{
		UserID: userID,
	}
}

func (i *IntentAddIdentity) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentAddIdentity) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []newinteraction.Edge{
			&nodes.EdgeCreateIdentityBegin{},
		}, nil

	case *nodes.NodeCreateIdentityEnd:
		return []newinteraction.Edge{
			&nodes.EdgeDoCreateIdentity{
				Identity: node.IdentityInfo,
			},
		}, nil
	case *nodes.NodeDoCreateIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{
				Stage: newinteraction.AuthenticationStagePrimary,
			},
		}, nil

	case *nodes.NodeCreateAuthenticatorEnd:
		return []newinteraction.Edge{
			&nodes.EdgeDoCreateAuthenticator{
				Stage:          node.Stage,
				Authenticators: node.Authenticators,
			},
		}, nil
	case *nodes.NodeDoCreateAuthenticator:
		switch node.Stage {
		case newinteraction.AuthenticationStagePrimary:
			return nil, nil
		default:
			panic("interaction: unexpected authenticator stage: " + node.Stage)
		}

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
