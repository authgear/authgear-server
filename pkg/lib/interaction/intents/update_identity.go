package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentUpdateIdentity{})
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

func (i *IntentUpdateIdentity) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentUpdateIdentity) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeUpdateIdentityBegin{
				IdentityID: i.IdentityID,
			},
		}, nil

	case *nodes.NodeUpdateIdentityEnd:
		return []interaction.Edge{
			&nodes.EdgeDoUpdateIdentity{
				IdentityBeforeUpdate: node.IdentityBeforeUpdate,
				IdentityAfterUpdate:  node.IdentityAfterUpdate,
			},
		}, nil
	case *nodes.NodeDoUpdateIdentity:
		return []interaction.Edge{
			&nodes.EdgeEnsureVerificationBegin{
				Identity:        node.IdentityAfterUpdate,
				RequestedByUser: false,
			},
		}, nil

	case *nodes.NodeEnsureVerificationEnd:
		return []interaction.Edge{
			&nodes.EdgeDoVerifyIdentity{
				Identity:         node.Identity,
				NewAuthenticator: node.NewAuthenticator,
			},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return []interaction.Edge{
			&nodes.EdgeDoUseIdentity{Identity: node.Identity},
		}, nil

	case *nodes.NodeDoUseIdentity:
		updateIdentity := mustFindNodeDoUpdateIdentity(graph)
		return []interaction.Edge{
			&nodes.EdgeRemoveAuthenticator{
				IdentityInfo: updateIdentity.IdentityBeforeUpdate,
			},
		}, nil

	case *nodes.NodeRemoveAuthenticator:
		return []interaction.Edge{
			&nodes.EdgeDoRemoveAuthenticator{
				Authenticators: node.Authenticators,
			},
		}, nil
	case *nodes.NodeDoRemoveAuthenticator:
		return []interaction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{
				Stage: interaction.AuthenticationStagePrimary,
			},
		}, nil

	case *nodes.NodeCreateAuthenticatorEnd:
		return []interaction.Edge{
			&nodes.EdgeDoCreateAuthenticator{
				Stage:          node.Stage,
				Authenticators: node.Authenticators,
			},
		}, nil
	case *nodes.NodeDoCreateAuthenticator:
		switch node.Stage {
		case interaction.AuthenticationStagePrimary:
			return nil, nil
		default:
			panic("interaction: unexpected authenticator stage: " + node.Stage)
		}

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
