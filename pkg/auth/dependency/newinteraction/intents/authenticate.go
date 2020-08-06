package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/errors"
)

func init() {
	newinteraction.RegisterIntent(&IntentAuthenticate{})
}

type IntentAuthenticateKind string

const (
	IntentAuthenticateKindLogin   IntentAuthenticateKind = "login"
	IntentAuthenticateKindSignup  IntentAuthenticateKind = "signup"
	IntentAuthenticateKindPromote IntentAuthenticateKind = "promote"
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

func NewIntentPromote() *IntentAuthenticate {
	return &IntentAuthenticate{Kind: IntentAuthenticateKindPromote}
}

func (i *IntentAuthenticate) InstantiateRootNode(ctx *newinteraction.Context, graph *newinteraction.Graph) (newinteraction.Node, error) {
	edge := nodes.EdgeSelectIdentityBegin{}
	return edge.Instantiate(ctx, graph, i)
}

// nolint:gocyclo
func (i *IntentAuthenticate) DeriveEdgesForNode(ctx *newinteraction.Context, graph *newinteraction.Graph, node newinteraction.Node) ([]newinteraction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeSelectIdentityEnd:
		switch i.Kind {
		case IntentAuthenticateKindLogin:
			if node.IdentityInfo == nil {
				switch node.IdentitySpec.Type {
				// Special case: login with new OAuth/anonymous identity means signup.
				case authn.IdentityTypeOAuth, authn.IdentityTypeAnonymous:
					return []newinteraction.Edge{
						&nodes.EdgeDoCreateUser{},
					}, nil
				default:
					return nil, newinteraction.ErrInvalidCredentials
				}
			}

			return []newinteraction.Edge{
				&nodes.EdgeDoUseIdentity{Identity: node.IdentityInfo},
			}, nil
		case IntentAuthenticateKindSignup:
			if node.IdentityInfo != nil {
				switch node.IdentitySpec.Type {
				// Special case: signup with existing OAuth identity means login.
				case authn.IdentityTypeOAuth:
					return []newinteraction.Edge{
						&nodes.EdgeDoUseIdentity{Identity: node.IdentityInfo},
					}, nil
				default:
					return nil, newinteraction.ErrDuplicatedIdentity
				}
			}

			return []newinteraction.Edge{
				&nodes.EdgeDoCreateUser{},
			}, nil
		case IntentAuthenticateKindPromote:
			if node.IdentityInfo == nil || node.IdentityInfo.Type != authn.IdentityTypeAnonymous {
				return nil, errors.New("promote intent is used to select non-anonymous identity")
			}

			return []newinteraction.Edge{
				&nodes.EdgeDoUseIdentity{Identity: node.IdentityInfo},
			}, nil
		default:
			panic("interaction: unknown authentication intent kind: " + i.Kind)
		}

	case *nodes.NodeDoCreateUser:
		selectIdentity := mustFindNodeSelectIdentity(graph)

		return []newinteraction.Edge{
			&nodes.EdgeCreateIdentityEnd{
				IdentitySpec: selectIdentity.IdentitySpec,
			},
		}, nil

	case *nodes.NodeCreateIdentityEnd:
		return []newinteraction.Edge{
			&nodes.EdgeCheckIdentityConflict{
				NewIdentity: node.IdentityInfo,
			},
		}, nil

	case *nodes.NodeCheckIdentityConflict:
		if node.DuplicatedIdentity == nil {
			return []newinteraction.Edge{
				&nodes.EdgeDoCreateIdentity{
					Identity: node.NewIdentity,
				},
			}, nil
		}

		switch i.Kind {
		case IntentAuthenticateKindPromote:
			switch ctx.Config.Identity.OnConflict.Promotion {
			case config.PromotionConflictBehaviorError:
				return nil, newinteraction.ErrDuplicatedIdentity
			case config.PromotionConflictBehaviorLogin:
				// Authenticate using duplicated identity
				return []newinteraction.Edge{
					&nodes.EdgeDoUseIdentity{
						Identity: node.DuplicatedIdentity,
					},
				}, nil
			default:
				panic("interaction: unknown promotion conflict behavior: " + ctx.Config.Identity.OnConflict.Promotion)
			}
		default:
			// TODO(interaction): handle OAuth identity conflict
			return nil, newinteraction.ErrDuplicatedIdentity
		}

	case *nodes.NodeDoCreateIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeEnsureVerificationBegin{
				Identity:        node.Identity,
				RequestedByUser: false,
			},
		}, nil

	case *nodes.NodeEnsureVerificationEnd:
		if node.NewAuthenticator != nil {
			return []newinteraction.Edge{
				&nodes.EdgeDoVerifyIdentity{
					Identity:         node.Identity,
					NewAuthenticator: node.NewAuthenticator,
				},
			}, nil
		}
		return []newinteraction.Edge{
			&nodes.EdgeDoUseIdentity{Identity: node.Identity},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return []newinteraction.Edge{
			&nodes.EdgeDoUseIdentity{Identity: node.Identity},
		}, nil

	case *nodes.NodeDoUseIdentity:
		if i.Kind == IntentAuthenticateKindPromote {
			if node.Identity.Type == authn.IdentityTypeAnonymous {
				// Create new identity for the anonymous user
				return []newinteraction.Edge{
					&nodes.EdgeCreateIdentityBegin{
						AllowAnonymousUser: false,
					},
				}, nil
			}

			selectIdentity := mustFindNodeSelectIdentity(graph)
			if selectIdentity.IdentityInfo.Type != authn.IdentityTypeAnonymous {
				panic("interaction: expect anonymous identity")
			}

			if selectIdentity.IdentityInfo.UserID == node.Identity.UserID {
				// Remove anonymous identity before proceeding
				return []newinteraction.Edge{
					&nodes.EdgeDoRemoveIdentity{
						Identity: selectIdentity.IdentityInfo,
					},
				}, nil
			}
		}

		_, isNewUser := graph.GetNewUserID()
		if isNewUser {
			// No authentication needed for new users
			return []newinteraction.Edge{
				&nodes.EdgeCreateAuthenticatorBegin{
					Stage: newinteraction.AuthenticationStagePrimary,
				},
			}, nil
		}
		return []newinteraction.Edge{
			&nodes.EdgeAuthenticationBegin{
				Stage: newinteraction.AuthenticationStagePrimary,
			},
		}, nil

	case *nodes.NodeDoRemoveIdentity:
		if node.Identity.Type != authn.IdentityTypeAnonymous {
			panic("interaction: expect anonymous identity")
		}

		// After removing anonymous identity:
		// continue to create authenticators
		return []newinteraction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{
				Stage: newinteraction.AuthenticationStagePrimary,
			},
		}, nil

	case *nodes.NodeAuthenticationEnd:
		switch node.Stage {
		case newinteraction.AuthenticationStagePrimary:
			return []newinteraction.Edge{
				&nodes.EdgeDoUseAuthenticator{
					Stage:         newinteraction.AuthenticationStagePrimary,
					Authenticator: node.VerifiedAuthenticator,
				},
			}, nil

		case newinteraction.AuthenticationStageSecondary:
			return []newinteraction.Edge{
				&nodes.EdgeDoUseAuthenticator{
					Stage:         newinteraction.AuthenticationStageSecondary,
					Authenticator: node.VerifiedAuthenticator,
				},
			}, nil

		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}
	case *nodes.NodeDoUseAuthenticator:
		switch node.Stage {
		case newinteraction.AuthenticationStagePrimary:
			return []newinteraction.Edge{
				&nodes.EdgeAuthenticationBegin{Stage: newinteraction.AuthenticationStageSecondary},
			}, nil

		case newinteraction.AuthenticationStageSecondary:
			return []newinteraction.Edge{
				&nodes.EdgeCreateAuthenticatorBegin{Stage: newinteraction.AuthenticationStagePrimary},
			}, nil

		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}

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
			return []newinteraction.Edge{
				&nodes.EdgeCreateAuthenticatorBegin{Stage: newinteraction.AuthenticationStageSecondary},
			}, nil

		case newinteraction.AuthenticationStageSecondary:
			var reason auth.SessionCreateReason
			_, ok := graph.GetNewUserID()
			if ok {
				reason = auth.SessionCreateReasonSignup
			} else {
				reason = auth.SessionCreateReasonLogin
			}

			return []newinteraction.Edge{
				&nodes.EdgeDoCreateSession{Reason: reason},
			}, nil

		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}

	case *nodes.NodeDoCreateSession:
		// Intent is finished
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
