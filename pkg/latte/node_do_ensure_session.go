package latte

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoEnsureSession{})
}

var _ workflow.CookieGetter = &NodeDoEnsureSession{}
var _ workflow.AuthenticationInfoEntryGetter = &NodeDoEnsureSession{}

type NodeDoEnsureSession struct {
	UserID                  string                    `json:"user_id"`
	CreateReason            session.CreateReason      `json:"create_reason"`
	SessionToCreate         *idpsession.IDPSession    `json:"session_to_create,omitempty"`
	AuthenticationInfoEntry *authenticationinfo.Entry `json:"authentication_info_entry,omitempty"`
	SessionCookie           *http.Cookie              `json:"session_cookie,omitempty"`
	UpdateSessionID         string                    `json:"update_session_id,omitempty"`
	UpdateSessionAMR        []string                  `json:"update_session_amr,omitempty"`
	SameSiteStrictCookie    *http.Cookie              `json:"same_site_strict_cookie,omitempty"`
}

func (n *NodeDoEnsureSession) Kind() string {
	return "latte.NodeDoEnsureSession"
}

func (n *NodeDoEnsureSession) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			return deps.AuthenticationInfos.Save(ctx, n.AuthenticationInfoEntry)
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			now := deps.Clock.NowUTC()
			return deps.Users.UpdateLoginTime(ctx, n.UserID, now)
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			if n.SessionToCreate == nil {
				return nil
			}

			err := deps.IDPSessions.Create(ctx, n.SessionToCreate)
			if err != nil {
				return err
			}

			s := session.GetSession(ctx)
			if s != nil && s.SessionType() == session.TypeIdentityProvider {
				err = deps.Sessions.RevokeWithoutEvent(ctx, s)
				if err != nil {
					return err
				}
			}

			return nil
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			if n.UpdateSessionID == "" {
				return nil
			}
			err = deps.IDPSessions.Reauthenticate(ctx, n.UpdateSessionID, n.UpdateSessionAMR)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (n *NodeDoEnsureSession) GetCookies(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	if n.SessionCookie != nil {
		cookies = append(cookies, n.SessionCookie)
	}
	if n.SameSiteStrictCookie != nil {
		cookies = append(cookies, n.SameSiteStrictCookie)
	}
	return cookies, nil
}

func (n *NodeDoEnsureSession) GetAuthenticationInfoEntry(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) *authenticationinfo.Entry {
	return n.AuthenticationInfoEntry
}

func (*NodeDoEnsureSession) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoEnsureSession) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeDoEnsureSession) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}
