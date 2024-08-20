package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

func init() {
	authflow.RegisterNode(&NodeDidReauthenticate{})
}

type NodeDidReauthenticate struct {
	UserID string `json:"user_id"`

	AuthenticationInfoEntry *authenticationinfo.Entry `json:"authentication_info_entry,omitempty"`
}

func NewNodeDidReauthenticate(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, n *NodeDidReauthenticate) (*NodeDidReauthenticate, error) {
	attrs := session.NewAttrs(n.UserID)
	amr, err := collectAMR(ctx, deps, flows)
	if err != nil {
		return nil, err
	}
	attrs.SetAMR(amr)
	authnInfo := authenticationinfo.T{
		UserID:          n.UserID,
		AuthenticatedAt: deps.Clock.NowUTC(),
		AMR:             amr,
	}
	authnInfoEntry := authenticationinfo.NewEntry(authnInfo,
		authflow.GetOAuthSessionID(ctx),
		authflow.GetSAMLSessionID(ctx),
	)

	n.AuthenticationInfoEntry = authnInfoEntry

	return n, nil
}

var _ authflow.NodeSimple = &NodeDidReauthenticate{}
var _ authflow.Milestone = &NodeDidReauthenticate{}
var _ MilestoneDidReauthenticate = &NodeDidReauthenticate{}
var _ authflow.EffectGetter = &NodeDidReauthenticate{}
var _ authflow.AuthenticationInfoEntryGetter = &NodeDidReauthenticate{}

func (*NodeDidReauthenticate) Kind() string {
	return "NodeDidReauthenticate"
}

func (*NodeDidReauthenticate) Milestone() {}
func (n *NodeDidReauthenticate) MilestoneDidReauthenticate() {
}

func (n *NodeDidReauthenticate) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			return deps.AuthenticationInfos.Save(n.AuthenticationInfoEntry)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			now := deps.Clock.NowUTC()
			return deps.Users.UpdateLoginTime(n.UserID, now)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			s := session.GetSession(ctx)
			if idp, ok := s.(*idpsession.IDPSession); ok && idp.GetUserID() == n.UserID {
				err = deps.IDPSessions.Reauthenticate(idp.ID, n.AuthenticationInfoEntry.T.AMR)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (n *NodeDidReauthenticate) GetAuthenticationInfoEntry(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) *authenticationinfo.Entry {
	return n.AuthenticationInfoEntry
}
