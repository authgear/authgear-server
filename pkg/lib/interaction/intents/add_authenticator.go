package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentAddAuthenticator{})
}

type IntentAddAuthenticator struct {
	UserID            string                          `json:"user_id"`
	Stage             interaction.AuthenticationStage `json:"stage"`
	AuthenticatorType authn.AuthenticatorType         `json:"authenticator_type"`
}

func NewIntentAddAuthenticator(userID string, stage interaction.AuthenticationStage, t authn.AuthenticatorType) *IntentAddAuthenticator {
	return &IntentAddAuthenticator{
		UserID:            userID,
		Stage:             stage,
		AuthenticatorType: t,
	}
}

func (i *IntentAddAuthenticator) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserID}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentAddAuthenticator) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{
				Stage:             i.Stage,
				AuthenticatorType: &i.AuthenticatorType,
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
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
