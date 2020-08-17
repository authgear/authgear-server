package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentAddIdentity{})
}

type IntentAddIdentity struct {
	UserID string `json:"user_id"`
}

func NewIntentAddIdentity(userID string) *IntentAddIdentity {
	return &IntentAddIdentity{
		UserID: userID,
	}
}

func (i *IntentAddIdentity) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentAddIdentity) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeCreateIdentityBegin{},
		}, nil

	case *nodes.NodeCreateIdentityEnd:
		return []interaction.Edge{
			&nodes.EdgeDoCreateIdentity{
				Identity: node.IdentityInfo,
			},
		}, nil
	case *nodes.NodeDoCreateIdentity:
		return []interaction.Edge{
			&nodes.EdgeEnsureVerificationBegin{
				Identity:        node.Identity,
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
