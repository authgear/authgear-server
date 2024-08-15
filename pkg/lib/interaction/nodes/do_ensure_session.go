package nodes

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

func init() {
	interaction.RegisterNode(&NodeDoEnsureSession{})
}

type EnsureSessionMode string

const (
	EnsureSessionModeDefault        EnsureSessionMode = ""
	EnsureSessionModeCreate         EnsureSessionMode = "create"
	EnsureSessionModeUpdateOrCreate EnsureSessionMode = "update_or_create"
	EnsureSessionModeNoop           EnsureSessionMode = "noop"
)

type EdgeDoEnsureSession struct {
	CreateReason session.CreateReason
	Mode         EnsureSessionMode
}

func (e *EdgeDoEnsureSession) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	amr := graph.GetAMR()
	userID := graph.MustGetUserID()

	mode := e.Mode
	if mode == EnsureSessionModeDefault {
		mode = EnsureSessionModeCreate
	}

	attrs := session.NewAttrs(userID)
	attrs.SetAMR(amr)
	var sessionToCreate *idpsession.IDPSession
	newSession, token := ctx.Sessions.MakeSession(attrs)
	sessionToCreate = newSession
	sessionCookie := ctx.CookieManager.ValueCookie(ctx.SessionCookie.Def, token)

	var updateSessionID string
	var updateSessionAMR []string
	if mode == EnsureSessionModeUpdateOrCreate {
		s := session.GetSession(ctx.Request.Context())
		if idp, ok := s.(*idpsession.IDPSession); ok && idp.GetUserID() == userID {
			updateSessionID = idp.ID
			updateSessionAMR = amr
			sessionToCreate = nil
			sessionCookie = nil
		}
	}

	if mode == EnsureSessionModeNoop {
		updateSessionID = ""
		updateSessionAMR = nil
		sessionToCreate = nil
		sessionCookie = nil
	}

	sameSiteStrictCookie := ctx.CookieManager.ValueCookie(
		ctx.SessionCookie.SameSiteStrictDef,
		"true",
	)

	now := ctx.Clock.NowUTC()

	var authenticationInfo authenticationinfo.T
	if sessionToCreate == nil {
		authenticationInfo = newSession.GetAuthenticationInfo()
	} else {
		authenticationInfo = sessionToCreate.CreateNewAuthenticationInfoByThisSession()
	}
	authenticationInfo.ShouldFireAuthenticatedEventWhenIssueOfflineGrant = mode == EnsureSessionModeNoop && e.CreateReason == session.CreateReasonLogin
	authenticationInfoEntry := authenticationinfo.NewEntry(authenticationInfo, ctx.OAuthSessionID, "")

	return &NodeDoEnsureSession{
		CreateReason:            e.CreateReason,
		SessionToCreate:         sessionToCreate,
		AuthenticationInfoEntry: authenticationInfoEntry,
		UpdateLoginTime:         now,
		UpdateSessionID:         updateSessionID,
		UpdateSessionAMR:        updateSessionAMR,
		SessionCookie:           sessionCookie,
		SameSiteStrictCookie:    sameSiteStrictCookie,
		IsAdminAPI:              interaction.IsAdminAPI(input),
	}, nil
}

type NodeDoEnsureSession struct {
	CreateReason            session.CreateReason      `json:"reason"`
	SessionToCreate         *idpsession.IDPSession    `json:"session_to_create,omitempty"`
	AuthenticationInfoEntry *authenticationinfo.Entry `json:"authentication_info_entry,omitempty"`
	UpdateLoginTime         time.Time                 `json:"update_login_time,omitempty"`
	UpdateSessionID         string                    `json:"update_session_id,omitempty"`
	UpdateSessionAMR        []string                  `json:"update_session_amr,omitempty"`
	SessionCookie           *http.Cookie              `json:"session_cookie,omitempty"`
	SameSiteStrictCookie    *http.Cookie              `json:"same_site_strict_cookie,omitempty"`
	IsAdminAPI              bool                      `json:"is_admin_api"`
}

// GetCookies implements CookiesGetter
func (n *NodeDoEnsureSession) GetCookies() (cookies []*http.Cookie) {
	if n.SessionCookie != nil {
		cookies = append(cookies, n.SessionCookie)
	}
	if n.SameSiteStrictCookie != nil {
		cookies = append(cookies, n.SameSiteStrictCookie)
	}
	return
}

func (n *NodeDoEnsureSession) GetAuthenticationInfoEntry() *authenticationinfo.Entry {
	return n.AuthenticationInfoEntry
}

func (n *NodeDoEnsureSession) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

// nolint:gocognit
func (n *NodeDoEnsureSession) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			return ctx.AuthenticationInfoService.Save(n.AuthenticationInfoEntry)
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.CreateReason != session.CreateReasonPromote {
				return nil
			}

			userID := graph.MustGetUserID()

			newUserRef := model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			}

			anonUserRef := newUserRef
			if identityCheck, ok := getIdentityConflictNode(graph); ok && identityCheck.DuplicatedIdentity != nil {
				// Logging as existing user when promoting: old user is different.
				anonUserRef = model.UserRef{
					Meta: model.Meta{
						ID: identityCheck.NewIdentity.UserID,
					},
				}
			}

			var identityModels []model.Identity
			for _, info := range graph.GetUserNewIdentities() {
				identityModels = append(identityModels, info.ToModel())
			}

			err := ctx.Events.DispatchEventOnCommit(&nonblocking.UserAnonymousPromotedEventPayload{
				AnonymousUserRef: anonUserRef,
				UserRef:          newUserRef,
				Identities:       identityModels,
				AdminAPI:         n.IsAdminAPI,
			})
			if err != nil {
				return err
			}

			return nil
		}),
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			userID := graph.MustGetUserID()

			var err error
			if !n.UpdateLoginTime.IsZero() {
				err = ctx.Users.UpdateLoginTime(userID, n.UpdateLoginTime)
				if err != nil {
					return err
				}
			}

			userRef := model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			}

			if n.SessionToCreate != nil {
				if n.CreateReason == session.CreateReasonLogin || n.CreateReason == session.CreateReasonReauthenticate {
					// ref: https://github.com/authgear/authgear-server/issues/2930
					// For authentication that involves IDP session will dispatch user.authenticated event here
					// For authentication that suppresses IDP session. e.g. biometric login
					// They are handled in their own node.
					err = ctx.Events.DispatchEventOnCommit(&nonblocking.UserAuthenticatedEventPayload{
						UserRef:  userRef,
						Session:  *n.SessionToCreate.ToAPIModel(),
						AdminAPI: n.IsAdminAPI,
					})
					if err != nil {
						return err
					}
				}
			}

			if n.SessionToCreate != nil {
				err = ctx.Sessions.Create(n.SessionToCreate)
				if err != nil {
					return err
				}

				// Clean up unreachable IdP Session.
				s := session.GetSession(ctx.Request.Context())
				if s != nil && s.SessionType() == session.TypeIdentityProvider {
					err = ctx.SessionManager.RevokeWithoutEvent(s)
					if err != nil {
						return err
					}
				}
			}

			if n.UpdateSessionID != "" {
				err = ctx.Sessions.Reauthenticate(n.UpdateSessionID, n.UpdateSessionAMR)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDoEnsureSession) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
