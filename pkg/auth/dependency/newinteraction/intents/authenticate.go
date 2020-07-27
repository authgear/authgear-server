package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterIntent(&IntentAuthenticate{})
}

type IntentAuthenticateKind string

const (
	IntentAuthenticateKindLogin  IntentAuthenticateKind = "login"
	IntentAuthenticateKindSignup IntentAuthenticateKind = "signup"
)

type IntentAuthenticate struct {
	Kind IntentAuthenticateKind `json:"kind"`
}

func NewIntentLogin() *IntentAuthenticate {
	return &IntentAuthenticate{Kind: IntentAuthenticateKindLogin}
}

func NewIntentSignup() *IntentAuthenticate {
	return &IntentAuthenticate{Kind: IntentAuthenticateKindSignup}
}

func (i *IntentAuthenticate) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	edge := nodes.EdgeSelectIdentityBegin{}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentAuthenticate) GetUseAnonymousUser() bool {
	return false
}

func (i *IntentAuthenticate) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeSelectIdentityEnd:
		switch i.Kind {
		case IntentAuthenticateKindLogin:
			if node.ExistingIdentity == nil {
				switch node.RequestedIdentity.Type {
				case authn.IdentityTypeOAuth, authn.IdentityTypeAnonymous:
					// Create new user with the requested identity if not exists
					return []newinteraction.Edge{
						&nodes.EdgeDoCreateUser{},
					}, nil

				default:
					return nil, newinteraction.ErrInvalidCredentials
				}
			}

			return []newinteraction.Edge{
				&nodes.EdgeAuthenticationBegin{
					Stage: firstAuthenticationStage(node.ExistingIdentity.Type),
				},
			}, nil

		case IntentAuthenticateKindSignup:
			if node.ExistingIdentity != nil {
				return nil, newinteraction.ErrDuplicatedIdentity
			}

			return []newinteraction.Edge{
				&nodes.EdgeDoCreateUser{},
			}, nil

		default:
			panic("interaction: unknown authentication intent kind: " + i.Kind)
		}

	case *nodes.NodeDoCreateUser:
		var selectIdentity *nodes.NodeSelectIdentityEnd
		for _, node := range graph.Nodes {
			if node, ok := node.(*nodes.NodeSelectIdentityEnd); ok {
				selectIdentity = node
				break
			}
		}
		if selectIdentity == nil {
			panic("interaction: expect identity already selected")
		}

		return []newinteraction.Edge{
			&nodes.EdgeCreateIdentityBegin{RequestedIdentity: selectIdentity.RequestedIdentity},
		}, nil

	case *nodes.NodeCreateIdentityEnd:
		return []newinteraction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{
				Stage: firstAuthenticationStage(graph.MustGetUserLastIdentity().Type),
			},
		}, nil

	case *nodes.NodeAuthenticationEnd:
		switch node.Stage {
		case newinteraction.AuthenticationStagePrimary:
			if node.Authenticator == nil {
				return nil, newinteraction.ErrInvalidCredentials
			}

			// TODO(interaction): check MFA mode
			return []newinteraction.Edge{
				&nodes.EdgeAuthenticationBegin{Stage: newinteraction.AuthenticationStageSecondary},
			}, nil

		case newinteraction.AuthenticationStageSecondary:
			return []newinteraction.Edge{
				&nodes.EdgeDoCreateSession{Reason: auth.SessionCreateReasonLogin},
			}, nil

		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}

	case *nodes.NodeCreateAuthenticatorEnd:
		switch node.Stage {
		case newinteraction.AuthenticationStagePrimary:
			// TODO(interaction): check MFA mode
			return []newinteraction.Edge{
				&nodes.EdgeCreateAuthenticatorBegin{Stage: newinteraction.AuthenticationStageSecondary},
			}, nil

		case newinteraction.AuthenticationStageSecondary:
			return []newinteraction.Edge{
				&nodes.EdgeDoCreateSession{Reason: auth.SessionCreateReasonSignup},
			}, nil

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
