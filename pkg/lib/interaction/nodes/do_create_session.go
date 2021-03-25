package nodes

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateSession{})
}

type EdgeDoCreateSession struct {
	Reason session.CreateReason
}

func (e *EdgeDoCreateSession) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	amr := graph.GetAMR()
	acr := graph.GetACR(amr)
	userIdentity := graph.MustGetUserLastIdentity()

	attrs := &session.Attrs{
		UserID: graph.MustGetUserID(),
		Claims: map[authn.ClaimName]interface{}{},
	}
	attrs.SetAMR(amr)
	attrs.SetACR(acr)
	if claimName, ok := userIdentity.DisplayIDClaimName(); ok {
		attrs.Claims[claimName] = userIdentity.DisplayID()
	}

	sess, token := ctx.Sessions.MakeSession(attrs)
	cookie := ctx.CookieFactory.ValueCookie(ctx.SessionCookie.Def, token)

	return &NodeDoCreateSession{
		Reason:        e.Reason,
		Session:       sess,
		SessionCookie: cookie,
	}, nil
}

type NodeDoCreateSession struct {
	Reason        session.CreateReason   `json:"reason"`
	Session       *idpsession.IDPSession `json:"session"`
	SessionCookie *http.Cookie           `json:"session_cookie"`
}

// GetCookies implements CookiesGetter
func (n *NodeDoCreateSession) GetCookies() []*http.Cookie {
	return []*http.Cookie{n.SessionCookie}
}

func (n *NodeDoCreateSession) SessionAttrs() *session.Attrs {
	return &n.Session.Attrs
}

func (n *NodeDoCreateSession) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeDoCreateSession) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.Reason != session.CreateReasonPromote {
				return nil
			}

			newUser, err := ctx.Users.Get(n.Session.Attrs.UserID)
			if err != nil {
				return err
			}

			anonUser := newUser
			if identityCheck, ok := getIdentityConflictNode(graph); ok && identityCheck.DuplicatedIdentity != nil {
				// Logging as existing user when promoting: old user is different.
				anonUser, err = ctx.Users.Get(identityCheck.NewIdentity.UserID)
				if err != nil {
					return err
				}
			}

			var identities []model.Identity
			for _, info := range graph.GetUserNewIdentities() {
				identities = append(identities, info.ToModel())
			}

			err = ctx.Hooks.DispatchEvent(&event.UserPromoteEvent{
				AnonymousUser: *anonUser,
				User:          *newUser,
				Identities:    identities,
			})
			if err != nil {
				return err
			}

			err = ctx.Hooks.DispatchEvent(&nonblocking.UserPromotedEvent{
				AnonymousUser: *anonUser,
				User:          *newUser,
				Identities:    identities,
			})
			if err != nil {
				return err
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			err := ctx.Users.UpdateLoginTime(n.Session.Attrs.UserID, n.Session.CreatedAt)
			if err != nil {
				return err
			}

			user, err := ctx.Users.Get(n.Session.Attrs.UserID)
			if err != nil {
				return err
			}

			err = ctx.Hooks.DispatchEvent(&event.SessionCreateEvent{
				Reason:  string(n.Reason),
				User:    *user,
				Session: *n.Session.ToAPIModel(),
			})
			if err != nil {
				return err
			}

			var e event.Payload
			switch n.Reason {
			case session.CreateReasonSignup:
				e = &nonblocking.SessionCreatedUserSignupEvent{
					User:    *user,
					Session: *n.Session.ToAPIModel(),
				}
			case session.CreateReasonLogin:
				e = &nonblocking.SessionCreatedUserLoginEvent{
					User:    *user,
					Session: *n.Session.ToAPIModel(),
				}
			case session.CreateReasonPromote:
				e = &nonblocking.SessionCreatedUserPromoteThemselvesEvent{
					User:    *user,
					Session: *n.Session.ToAPIModel(),
				}
			default:
				panic(fmt.Errorf("interaction: unexpected session create reason: %s", n.Reason))
			}

			err = ctx.Hooks.DispatchEvent(e)
			if err != nil {
				return err
			}

			err = ctx.Sessions.Create(n.Session)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoCreateSession) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
