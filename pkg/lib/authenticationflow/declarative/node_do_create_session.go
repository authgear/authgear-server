package declarative

import (
	"context"
	"net/http"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

func init() {
	authflow.RegisterNode(&NodeDoCreateSession{})
}

type NodeDoCreateSession struct {
	UserID       string               `json:"user_id"`
	CreateReason session.CreateReason `json:"create_reason"`
	SkipCreate   bool                 `json:"skip_create"`

	Session                 *idpsession.IDPSession    `json:"session,omitempty"`
	SessionCookie           *http.Cookie              `json:"session_cookie,omitempty"`
	AuthenticationInfoEntry *authenticationinfo.Entry `json:"authentication_info_entry,omitempty"`
	SameSiteStrictCookie    *http.Cookie              `json:"same_site_strict_cookie,omitempty"`
}

func NewNodeDoCreateSession(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoCreateSession) (*NodeDoCreateSession, error) {
	attrs := session.NewAttrs(n.UserID)
	amr, err := collectAMR(ctx, deps, flows)
	if err != nil {
		return nil, err
	}
	attrs.SetAMR(amr)
	s, token := deps.IDPSessions.MakeSession(attrs)
	sessionCookie := deps.Cookies.ValueCookie(deps.SessionCookie.Def, token)

	authnInfo := s.GetAuthenticationInfoByThisSession()
	authnInfo.ShouldFireAuthenticatedEventWhenIssueOfflineGrant = n.SkipCreate && n.CreateReason == session.CreateReasonLogin
	authnInfoEntry := authenticationinfo.NewEntry(authnInfo, authflow.GetOAuthSessionID(ctx))

	sameSiteStrictCookie := deps.Cookies.ValueCookie(
		deps.SessionCookie.SameSiteStrictDef,
		"true",
	)

	if n.SkipCreate {
		s = nil
		sessionCookie = nil
	}

	n.Session = s
	n.SessionCookie = sessionCookie
	n.AuthenticationInfoEntry = authnInfoEntry
	n.SameSiteStrictCookie = sameSiteStrictCookie

	return n, nil
}

var _ authflow.NodeSimple = &NodeDoCreateSession{}
var _ authflow.Milestone = &NodeDoCreateSession{}
var _ MilestoneDoCreateSession = &NodeDoCreateSession{}
var _ authflow.EffectGetter = &NodeDoCreateSession{}
var _ authflow.CookieGetter = &NodeDoCreateSession{}
var _ authflow.AuthenticationInfoEntryGetter = &NodeDoCreateSession{}

func (*NodeDoCreateSession) Kind() string {
	return "NodeDoCreateSession"
}

func (*NodeDoCreateSession) Milestone() {}
func (n *NodeDoCreateSession) MilestoneDoCreateSession() (*idpsession.IDPSession, bool) {
	if n.Session != nil {
		return n.Session, true
	}

	return nil, false
}

func (n *NodeDoCreateSession) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.AuthenticationInfos.Save(n.AuthenticationInfoEntry)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			now := deps.Clock.NowUTC()
			return deps.Users.UpdateLoginTime(n.UserID, now)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.Session == nil {
				return nil
			}

			err := deps.IDPSessions.Create(n.Session)
			if err != nil {
				return err
			}

			// Clean up unreachable IdP session
			s := session.GetSession(ctx)
			if s != nil && s.SessionType() == session.TypeIdentityProvider {
				err = deps.Sessions.RevokeWithoutEvent(s)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoCreateSession) GetCookies(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	if n.SessionCookie != nil {
		cookies = append(cookies, n.SessionCookie)
	}
	if n.SameSiteStrictCookie != nil {
		cookies = append(cookies, n.SameSiteStrictCookie)
	}
	return cookies, nil
}

func (n *NodeDoCreateSession) GetAuthenticationInfoEntry(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) *authenticationinfo.Entry {
	return n.AuthenticationInfoEntry
}
