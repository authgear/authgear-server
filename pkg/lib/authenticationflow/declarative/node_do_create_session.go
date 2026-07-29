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

	// ContinueFromSessionType and ContinueFromSessionID identify the
	// existing session that was reused instead of creating a new one (see
	// MilestoneDoUseExistingSession) — populated by the caller only when
	// SkipCreate is true because of that milestone, not merely because
	// x_suppress_idp_session_cookie was requested.
	ContinueFromSessionType session.Type `json:"continue_from_session_type,omitempty"`
	ContinueFromSessionID   string       `json:"continue_from_session_id,omitempty"`

	Session                 *idpsession.IDPSession    `json:"session,omitempty"`
	SessionCookie           *http.Cookie              `json:"session_cookie,omitempty"`
	AuthenticationInfoEntry *authenticationinfo.Entry `json:"authentication_info_entry,omitempty"`
	SameSiteStrictCookie    *http.Cookie              `json:"same_site_strict_cookie,omitempty"`
}

func NewNodeDoCreateSession(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDoCreateSession) (*NodeDoCreateSession, error) {
	identitySpecs, err := collectIdentitySpecs(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	var authnInfo authenticationinfo.T
	var newSession *idpsession.IDPSession = nil
	var sessionCookie *http.Cookie = nil

	if n.ContinueFromSessionID != "" {
		// The session being continued was already re-verified against the
		// current request by resolveSelectAccountSession, but it may have
		// been resolved in an earlier request than this one, so check again
		// here that it is still the same session before reusing its
		// authentication details as-is.
		sess := session.GetSession(ctx)
		if sess == nil || sess.SessionID() != n.ContinueFromSessionID || sess.SessionType() != n.ContinueFromSessionType {
			return nil, ErrSelectAccountSessionChanged
		}
		authnInfo = sess.CreateNewAuthenticationInfoByThisSession()
		authnInfo.IdentitySpecs = identitySpecs
		authnInfo.ContinueFromSessionType = string(n.ContinueFromSessionType)
		authnInfo.ContinueFromSessionID = n.ContinueFromSessionID
	} else {
		amr, err := CollectAMR(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		authnInfo = authenticationinfo.T{
			UserID:          n.UserID,
			AuthenticatedAt: deps.Clock.NowUTC(),
			AMR:             amr,
			IdentitySpecs:   identitySpecs,
		}

		if !n.SkipCreate {
			attrs := session.NewAttrs(n.UserID)
			attrs.SetAMR(amr)
			s, token := deps.IDPSessions.MakeSession(attrs)
			newSession = s
			sessionCookie = deps.Cookies.ValueCookie(deps.SessionCookie.Def, token)
			authnInfo.AuthenticatedBySessionID = newSession.SessionID()
			authnInfo.AuthenticatedBySessionType = string(newSession.SessionType())
		}
	}

	authnInfo.ShouldFireAuthenticatedEventWhenIssueOfflineGrant = n.SkipCreate && n.CreateReason == session.CreateReasonLogin

	sameSiteStrictCookie := deps.Cookies.ValueCookie(
		deps.SessionCookie.SameSiteStrictDef,
		"true",
	)

	authnInfoEntry := authenticationinfo.NewEntry(authnInfo,
		authflow.GetOAuthSessionID(ctx),
		authflow.GetSAMLSessionID(ctx),
	)

	n.Session = newSession
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
			return deps.AuthenticationInfos.Save(ctx, n.AuthenticationInfoEntry)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.Session == nil {
				return nil
			}

			err := deps.IDPSessions.Create(ctx, n.Session)
			if err != nil {
				return err
			}

			// Clean up unreachable IdP session
			s := session.GetSession(ctx)
			if s != nil && s.SessionType() == session.TypeIdentityProvider {
				err = deps.Sessions.RevokeWithoutEvent(ctx, s)
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
