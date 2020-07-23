package intents

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type IntentLogin struct {
	UseAnonymousUser bool `json:"use_anonymous_user"`
}

func (i *IntentLogin) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	spec := nodes.EdgeSelectIdentityBegin{}
	return spec.Instantiate(ctx, graph, i)
}

func (i *IntentLogin) GetUseAnonymousUser() bool {
	return i.UseAnonymousUser
}

func (i *IntentLogin) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeSelectIdentityEnd:
		// Ensure identity exists before performing authentication
		if node.ExistingIdentity == nil {
			if node.RequestedIdentity.Type == authn.IdentityTypeOAuth {
				panic("TODO(new_interaction): create new user & identity if not exist")
			}

			return nil, newinteraction.ErrInvalidCredentials
		}

		return []newinteraction.Edge{
			&nodes.EdgeAuthenticationBegin{Stage: newinteraction.AuthenticationStagePrimary},
		}, nil

	default:
		panic("interaction: unexpected node")
	}
}
