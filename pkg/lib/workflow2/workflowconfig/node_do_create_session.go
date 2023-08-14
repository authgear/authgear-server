package workflowconfig

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeDoCreateSession{})
}

type NodeDoCreateSession struct {
	UserID       string               `json:"user_id"`
	CreateReason session.CreateReason `json:"create_reason"`
	SkipCreate   bool                 `json:"skip_create"`

	Session                  *idpsession.IDPSession    `json:"session,omitempty"`
	SessionCookie            *http.Cookie              `json:"session_cookie,omitempty"`
	AuthenticationInfoEntry  *authenticationinfo.Entry `json:"authentication_info_entry,omitempty"`
	AuthenticationInfoCookie *http.Cookie              `json:"authentication_info_cookie,omitempty"`
	SameSiteStrictCookie     *http.Cookie              `json:"same_site_strict_cookie,omitempty"`
}

var _ MilestoneDoCreateSession = &NodeDoCreateSession{}

func (*NodeDoCreateSession) Milestone() {}
func (n *NodeDoCreateSession) MilestoneDoCreateSession() (*idpsession.IDPSession, bool) {
	if n.Session != nil {
		return n.Session, true
	}

	return nil, false
}

func NewNodeDoCreateSession(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, n *NodeDoCreateSession) (*NodeDoCreateSession, error) {
	attrs := session.NewAttrs(n.UserID)
	amr, err := collectAMR(ctx, deps, workflows)
	if err != nil {
		return nil, err
	}
	attrs.SetAMR(amr)
	s, token := deps.IDPSessions.MakeSession(attrs)
	sessionCookie := deps.Cookies.ValueCookie(deps.SessionCookie.Def, token)

	authnInfo := s.GetAuthenticationInfo()
	authnInfo.ShouldFireAuthenticatedEventWhenIssueOfflineGrant = n.SkipCreate && n.CreateReason == session.CreateReasonLogin
	authnInfoEntry := authenticationinfo.NewEntry(authnInfo)
	authnInfoCookie := deps.Cookies.ValueCookie(
		authenticationinfo.CookieDef,
		authnInfoEntry.ID,
	)

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
	n.AuthenticationInfoCookie = authnInfoCookie
	n.SameSiteStrictCookie = sameSiteStrictCookie

	return n, nil
}

var _ workflow.NodeSimple = &NodeDoCreateSession{}
var _ workflow.EffectGetter = &NodeDoCreateSession{}
var _ workflow.CookieGetter = &NodeDoCreateSession{}

func (*NodeDoCreateSession) Kind() string {
	return "workflowconfig.NodeDoCreateSession"
}

func (n *NodeDoCreateSession) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.AuthenticationInfos.Save(n.AuthenticationInfoEntry)
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			now := deps.Clock.NowUTC()
			return deps.Users.UpdateLoginTime(n.UserID, now)
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
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

func (n *NodeDoCreateSession) GetCookies(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	if n.SessionCookie != nil {
		cookies = append(cookies, n.SessionCookie)
	}
	if n.SameSiteStrictCookie != nil {
		cookies = append(cookies, n.SameSiteStrictCookie)
	}
	if n.AuthenticationInfoCookie != nil {
		cookies = append(cookies, n.AuthenticationInfoCookie)
	}
	return cookies, nil
}
