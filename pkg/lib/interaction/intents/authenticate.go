package intents

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func init() {
	interaction.RegisterIntent(&IntentAuthenticate{})
}

type IntentAuthenticateKind string

const (
	IntentAuthenticateKindLogin   IntentAuthenticateKind = "login"
	IntentAuthenticateKindSignup  IntentAuthenticateKind = "signup"
	IntentAuthenticateKindPromote IntentAuthenticateKind = "promote"
)

type IntentAuthenticate struct {
	Kind              IntentAuthenticateKind `json:"kind"`
	SkipCreateSession bool                   `json:"skip_create_session"`
	WebhookState      string                 `json:"webhook_state"`
}

func NewIntentLogin(skipCreateSession bool) *IntentAuthenticate {
	return &IntentAuthenticate{
		Kind:              IntentAuthenticateKindLogin,
		SkipCreateSession: skipCreateSession,
	}
}

func NewIntentSignup(webhookState string) *IntentAuthenticate {
	return &IntentAuthenticate{
		Kind:              IntentAuthenticateKindSignup,
		SkipCreateSession: false,
		WebhookState:      webhookState,
	}
}

func NewIntentPromote() *IntentAuthenticate {
	return &IntentAuthenticate{
		Kind:              IntentAuthenticateKindPromote,
		SkipCreateSession: false,
	}
}

func (i *IntentAuthenticate) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	isAuthentication := i.Kind == IntentAuthenticateKindLogin
	edge := nodes.EdgeSelectIdentityBegin{
		IsAuthentication: isAuthentication,
	}
	return edge.Instantiate(ctx, graph, i)
}

// nolint:gocyclo
func (i *IntentAuthenticate) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeSelectIdentityEnd:
		switch i.Kind {
		case IntentAuthenticateKindLogin:
			if node.IdentityInfo == nil {
				switch node.IdentitySpec.Type {
				// Special case: login with new OAuth/anonymous identity means signup.
				case authn.IdentityTypeOAuth, authn.IdentityTypeAnonymous:
					return []interaction.Edge{
						&nodes.EdgeDoCreateUser{},
					}, nil
				default:
					return nil, interaction.ErrInvalidCredentials
				}
			}

			return []interaction.Edge{
				&nodes.EdgeEnsureVerificationBegin{
					Identity:        node.IdentityInfo,
					RequestedByUser: false,
				},
			}, nil
		case IntentAuthenticateKindSignup:
			if node.IdentityInfo != nil {
				switch node.IdentitySpec.Type {
				// Special case: signup with existing OAuth identity means login.
				case authn.IdentityTypeOAuth:
					return []interaction.Edge{
						&nodes.EdgeDoUseIdentity{Identity: node.IdentityInfo},
					}, nil
				default:
					return nil, node.IdentityInfo.FillDetails(interaction.ErrDuplicatedIdentity)
				}
			}

			return []interaction.Edge{
				&nodes.EdgeDoCreateUser{},
			}, nil
		case IntentAuthenticateKindPromote:
			if node.IdentityInfo == nil || node.IdentityInfo.Type != authn.IdentityTypeAnonymous {
				return nil, errors.New("promote intent is used to select non-anonymous identity")
			}

			return []interaction.Edge{
				&nodes.EdgeEnsureVerificationBegin{
					Identity:        node.IdentityInfo,
					RequestedByUser: false,
				},
			}, nil
		default:
			panic("interaction: unknown authentication intent kind: " + i.Kind)
		}

	case *nodes.NodeDoCreateUser:
		selectIdentity := mustFindNodeSelectIdentity(graph)

		return []interaction.Edge{
			&nodes.EdgeCreateIdentityEnd{
				IdentitySpec: selectIdentity.IdentitySpec,
			},
		}, nil

	case *nodes.NodeCreateIdentityEnd:
		return []interaction.Edge{
			&nodes.EdgeCheckIdentityConflict{
				NewIdentity: node.IdentityInfo,
			},
		}, nil

	case *nodes.NodeCheckIdentityConflict:
		if node.DuplicatedIdentity == nil {
			return []interaction.Edge{
				&nodes.EdgeDoCreateIdentity{
					Identity: node.NewIdentity,
				},
			}, nil
		}

		switch i.Kind {
		case IntentAuthenticateKindPromote:
			switch node.IdentityConflictConfig.Promotion {
			case config.PromotionConflictBehaviorError:
				return nil, node.DuplicatedIdentity.FillDetails(interaction.ErrDuplicatedIdentity)
			case config.PromotionConflictBehaviorLogin:
				// Authenticate using duplicated identity
				return []interaction.Edge{
					&nodes.EdgeDoUseIdentity{
						Identity: node.DuplicatedIdentity,
					},
				}, nil
			default:
				panic("interaction: unknown promotion conflict behavior: " + node.IdentityConflictConfig.Promotion)
			}
		default:
			// TODO(interaction): handle OAuth identity conflict
			return nil, node.DuplicatedIdentity.FillDetails(interaction.ErrDuplicatedIdentity)
		}

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
				NewVerifiedClaim: node.NewVerifiedClaim,
			},
		}, nil

	case *nodes.NodeDoVerifyIdentity:
		return []interaction.Edge{
			&nodes.EdgeDoUseIdentity{Identity: node.Identity},
		}, nil

	case *nodes.NodeDoUseIdentity:
		if i.Kind == IntentAuthenticateKindPromote {
			if node.Identity.Type == authn.IdentityTypeAnonymous {
				// Create new identity for the anonymous user
				return []interaction.Edge{
					&nodes.EdgeCreateIdentityBegin{},
				}, nil
			}

			selectIdentity := mustFindNodeSelectIdentity(graph)
			if selectIdentity.IdentityInfo.Type != authn.IdentityTypeAnonymous {
				panic("interaction: expect anonymous identity")
			}

			if selectIdentity.IdentityInfo.UserID == node.Identity.UserID {
				// Remove anonymous identity before proceeding
				return []interaction.Edge{
					&nodes.EdgeDoRemoveIdentity{
						Identity: selectIdentity.IdentityInfo,
					},
				}, nil
			}
		}

		_, isNewUser := graph.GetNewUserID()
		if isNewUser {
			// No authentication needed for new users
			return []interaction.Edge{
				&nodes.EdgeValidateUser{},
			}, nil
		}
		return []interaction.Edge{
			&nodes.EdgeAuthenticationBegin{
				Stage: authn.AuthenticationStagePrimary,
			},
		}, nil

	case *nodes.NodeDoRemoveIdentity:
		if node.Identity.Type != authn.IdentityTypeAnonymous {
			panic("interaction: expect anonymous identity")
		}

		// After removing anonymous identity:
		// continue to create authenticators (after validating user).
		return []interaction.Edge{
			&nodes.EdgeValidateUser{},
		}, nil

	case *nodes.NodeAuthenticationEnd:
		switch node.Stage {
		case authn.AuthenticationStagePrimary:
			return []interaction.Edge{
				&nodes.EdgeDoUseAuthenticator{
					Stage:         authn.AuthenticationStagePrimary,
					Authenticator: node.VerifiedAuthenticator,
				},
			}, nil

		case authn.AuthenticationStageSecondary:
			return []interaction.Edge{
				&nodes.EdgeDoUseAuthenticator{
					Stage:         authn.AuthenticationStageSecondary,
					Authenticator: node.VerifiedAuthenticator,
				},
			}, nil

		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}
	case *nodes.NodeDoUseAuthenticator:
		switch node.Stage {
		case authn.AuthenticationStagePrimary:
			return []interaction.Edge{
				&nodes.EdgeAuthenticationBegin{Stage: authn.AuthenticationStageSecondary},
			}, nil

		case authn.AuthenticationStageSecondary:
			return []interaction.Edge{
				&nodes.EdgeValidateUser{},
			}, nil

		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}

	case *nodes.NodeValidateUser:
		if node.Error != nil {
			// Stop interaction if user is invalid.
			return []interaction.Edge{
				&nodes.EdgeTerminal{},
			}, nil
		}

		return []interaction.Edge{
			&nodes.EdgeCreateAuthenticatorBegin{Stage: authn.AuthenticationStagePrimary},
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
		case authn.AuthenticationStagePrimary:
			return []interaction.Edge{
				&nodes.EdgeCreateAuthenticatorBegin{Stage: authn.AuthenticationStageSecondary},
			}, nil

		case authn.AuthenticationStageSecondary:
			return []interaction.Edge{
				&nodes.EdgeGenerateRecoveryCode{},
			}, nil

		default:
			panic(fmt.Errorf("interaction: unexpected authentication stage: %v", node.Stage))
		}

	case *nodes.NodeGenerateRecoveryCodeEnd:
		return []interaction.Edge{
			&nodes.EdgeDoGenerateRecoveryCode{
				RecoveryCodes: node.RecoveryCodes,
			},
		}, nil
	case *nodes.NodeDoGenerateRecoveryCode:
		var reason session.CreateReason
		_, creating := graph.GetNewUserID()
		switch {
		case i.Kind == IntentAuthenticateKindPromote:
			reason = session.CreateReasonPromote
		case creating:
			reason = session.CreateReasonSignup
		default:
			reason = session.CreateReasonLogin
		}

		return []interaction.Edge{
			&nodes.EdgeDoCreateSession{
				Reason:            reason,
				SkipCreateSession: i.SkipCreateSession,
			},
		}, nil

	case *nodes.NodeDoCreateSession:
		// Intent is finished
		return nil, nil

	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}

func (i *IntentAuthenticate) GetWebhookState() string {
	return i.WebhookState
}

var _ interaction.IntentWithWebhookState = &IntentAuthenticate{}
