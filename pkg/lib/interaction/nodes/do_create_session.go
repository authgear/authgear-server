package nodes

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

func init() {
	interaction.RegisterNode(&NodeDoCreateSession{})
}

type EdgeDoCreateSession struct {
	Reason            session.CreateReason
	SkipCreateSession bool
}

func (e *EdgeDoCreateSession) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	amr := graph.GetAMR()
	acr := graph.GetACR(amr)
	userIdentity := graph.MustGetUserLastIdentity()

	attrs := session.NewAttrs(graph.MustGetUserID())
	attrs.SetAMR(amr)
	attrs.SetACR(acr)
	if claimName, ok := userIdentity.DisplayIDClaimName(); ok {
		attrs.Claims[claimName] = userIdentity.DisplayID()
	}

	sess, token := ctx.Sessions.MakeSession(attrs)
	cookie := ctx.CookieFactory.ValueCookie(ctx.SessionCookie.Def, token)

	return &NodeDoCreateSession{
		Reason:            e.Reason,
		SkipCreateSession: e.SkipCreateSession,
		Session:           sess,
		SessionCookie:     cookie,
		IsAdminAPI:        interaction.IsAdminAPI(input),
	}, nil
}

type NodeDoCreateSession struct {
	Reason            session.CreateReason   `json:"reason"`
	SkipCreateSession bool                   `json:"skip_create_session"`
	Session           *idpsession.IDPSession `json:"session"`
	SessionCookie     *http.Cookie           `json:"session_cookie"`
	IsAdminAPI        bool                   `json:"is_admin_api"`
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

			err = ctx.Hooks.DispatchEvent(&nonblocking.UserAnonymousPromotedEvent{
				AnonymousUser: *anonUser,
				User:          *newUser,
				Identities:    identities,
				AdminAPI:      n.IsAdminAPI,
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

			if n.Reason == session.CreateReasonLogin {
				err = ctx.Hooks.DispatchEvent(&nonblocking.UserAuthenticatedEvent{
					User:     *user,
					Session:  *n.Session.ToAPIModel(),
					AdminAPI: n.IsAdminAPI,
				})
				if err != nil {
					return err
				}
			}

			if !n.SkipCreateSession {
				err = ctx.Sessions.Create(n.Session)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoCreateSession) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
