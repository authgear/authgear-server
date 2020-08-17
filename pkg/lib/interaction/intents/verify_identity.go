package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
)

func init() {
	interaction.RegisterIntent(&IntentVerifyIdentity{})
}

type IntentVerifyIdentity struct {
	UserID       string             `json:"user_id"`
	IdentityType authn.IdentityType `json:"identity_type"`
	IdentityID   string             `json:"identity_id"`
}

func NewIntentVerifyIdentity(userID string, identityType authn.IdentityType, identityID string) *IntentVerifyIdentity {
	return &IntentVerifyIdentity{
		UserID:       userID,
		IdentityType: identityType,
		IdentityID:   identityID,
	}
}

func (i *IntentVerifyIdentity) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	identityInfo, err := ctx.Identities.Get(i.UserID, i.IdentityType, i.IdentityID)
	if err != nil {
		return nil, err
	}
	return &nodes.NodeEnsureVerificationBegin{
		Identity:        identityInfo,
		RequestedByUser: true,
	}, nil
}

func (i *IntentVerifyIdentity) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeEnsureVerificationEnd:
		return []interaction.Edge{
			&nodes.EdgeDoVerifyIdentity{
				Identity:         node.Identity,
				NewAuthenticator: node.NewAuthenticator,
			},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
