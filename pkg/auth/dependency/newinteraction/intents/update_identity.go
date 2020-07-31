package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
)

func init() {
	newinteraction.RegisterIntent(&IntentUpdateIdentity{})
}

type IntentUpdateIdentity struct {
	UserID     string `json:"user_id"`
	IdentityID string `json:"identity_id"`
}

func NewIntentUpdateIdentity(userID string, identityID string) *IntentUpdateIdentity {
	return &IntentUpdateIdentity{
		UserID:     userID,
		IdentityID: identityID,
	}
}

func (i *IntentUpdateIdentity) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	edge := nodes.EdgeUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentUpdateIdentity) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeUseUser:
		return []newinteraction.Edge{
			&nodes.EdgeUpdateIdentityBegin{
				IdentityID: i.IdentityID,
			},
		}, nil
	case *nodes.NodeRemoveIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeRemoveAuthenticator{
				IdentityInfo: node.IdentityInfo,
			},
		}, nil
	case *nodes.NodeRemoveAuthenticator:
		return nil, nil
	case *nodes.NodeUpdateIdentityEnd:
		return []newinteraction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{
				Stage: newinteraction.AuthenticationStagePrimary,
			},
		}, nil
	case *nodes.NodeCreateAuthenticatorEnd:
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
