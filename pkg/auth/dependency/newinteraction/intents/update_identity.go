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
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentUpdateIdentity) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []newinteraction.Edge{
			&nodes.EdgeUpdateIdentityBegin{
				IdentityID: i.IdentityID,
			},
		}, nil

	case *nodes.NodeUpdateIdentityEnd:
		return []newinteraction.Edge{
			&nodes.EdgeDoUpdateIdentity{
				IdentityBeforeUpdate: node.IdentityBeforeUpdate,
				IdentityAfterUpdate:  node.IdentityAfterUpdate,
			},
		}, nil
	case *nodes.NodeDoUpdateIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeEnsureVerificationBegin{
				Identity:        node.IdentityAfterUpdate,
				RequestedByUser: false,
			},
		}, nil

	case *nodes.NodeEnsureVerificationEnd:
		return []newinteraction.Edge{
			&nodes.EdgeDoVerifyIdentity{
				Identity:         node.Identity,
				NewAuthenticator: node.NewAuthenticator,
			},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeDoUseIdentity{Identity: node.Identity},
		}, nil

	case *nodes.NodeDoUseIdentity:
		updateIdentity := mustFindNodeDoUpdateIdentity(graph)
		return []newinteraction.Edge{
			&nodes.EdgeRemoveAuthenticator{
				IdentityInfo: updateIdentity.IdentityBeforeUpdate,
			},
		}, nil

	case *nodes.NodeRemoveAuthenticator:
		return []newinteraction.Edge{
			&nodes.EdgeDoRemoveAuthenticator{
				Authenticators: node.Authenticators,
			},
		}, nil
	case *nodes.NodeDoRemoveAuthenticator:
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
