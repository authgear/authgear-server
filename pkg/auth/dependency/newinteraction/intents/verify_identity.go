package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterIntent(&IntentVerifyIdentity{})
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

func (i *IntentVerifyIdentity) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	identityInfo, err := ctx.Identities.Get(i.UserID, i.IdentityType, i.IdentityID)
	if err != nil {
		return nil, err
	}
	return &nodes.NodeEnsureVerificationBegin{
		Identity:        identityInfo,
		RequestedByUser: true,
	}, nil
}

func (i *IntentVerifyIdentity) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeEnsureVerificationEnd:
		return []newinteraction.Edge{
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
