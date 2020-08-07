package nodes

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeDoCreateSession{})
}

type EdgeDoCreateSession struct {
	Reason auth.SessionCreateReason
}

func (e *EdgeDoCreateSession) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	amr := graph.GetAMR()
	acr := graph.GetACR(amr)
	userIdentity := graph.MustGetUserLastIdentity()

	attrs := &authn.Attrs{
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
	Reason        auth.SessionCreateReason `json:"reason"`
	Session       *session.IDPSession      `json:"session"`
	SessionCookie *http.Cookie             `json:"session_cookie"`
}

// GetCookies implements CookiesGetter
func (n *NodeDoCreateSession) GetCookies() []*http.Cookie {
	return []*http.Cookie{n.SessionCookie}
}

func (n *NodeDoCreateSession) AuthnAttrs() authn.Attrs {
	return n.Session.Attrs
}

func (n *NodeDoCreateSession) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	err := perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		if n.Reason != auth.SessionCreateReasonPromote {
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

		err = ctx.Hooks.DispatchEvent(
			event.UserPromoteEvent{
				AnonymousUser: *anonUser,
				User:          *newUser,
				Identities:    identities,
			},
			newUser,
		)
		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	err = perform(newinteraction.EffectOnCommit(func(ctx *newinteraction.Context) error {
		user, err := ctx.Users.Get(n.Session.Attrs.UserID)
		if err != nil {
			return err
		}

		err = ctx.Users.UpdateLoginTime(user, n.Session.CreatedAt)
		if err != nil {
			return err
		}

		err = ctx.Hooks.DispatchEvent(
			event.SessionCreateEvent{
				Reason:  string(n.Reason),
				User:    *user,
				Session: *n.Session.ToAPIModel(),
			},
			user,
		)
		if err != nil {
			return err
		}

		err = ctx.Sessions.Create(n.Session)
		if err != nil {
			return err
		}

		return nil
	}))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeDoCreateSession) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(ctx, graph, n)
}
