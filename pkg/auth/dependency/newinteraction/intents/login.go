package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterIntent(&IntentLogin{})
}

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

	case *nodes.NodeAuthenticationEnd:
		switch node.Stage {
		case newinteraction.AuthenticationStagePrimary:
			if node.Authenticator == nil {
				identityType := graph.MustGetUserLastIdentity().Type
				switch identityType {
				case authn.IdentityTypeLoginID:
					return nil, newinteraction.ErrInvalidCredentials

				case authn.IdentityTypeAnonymous, authn.IdentityTypeOAuth:
					// Primary authentication is not needed
					break

				default:
					panic("interaction: unknown identity type: " + identityType)
				}
			}
			return []newinteraction.Edge{
				&nodes.EdgeAuthenticationBegin{Stage: newinteraction.AuthenticationStageSecondary},
			}, nil
		case newinteraction.AuthenticationStageSecondary:
			// TODO(new_interaction): MFA")
			return []newinteraction.Edge{&nodes.EdgeDoCreateSession{}}, nil
		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}
	case *nodes.NodeDoCreateSession:
		// Intent is finished
		return nil, nil

	default:
		panic("interaction: unexpected node")
	}
}
