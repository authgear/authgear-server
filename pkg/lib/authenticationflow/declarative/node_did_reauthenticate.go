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
	amr, err := CollectAMR(ctx, deps, flows)
	if err != nil {
		return nil, err
	}
	attrs.SetAMR(amr)

	identitySpecs, err := collectIdentitySpecs(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	authnInfo := authenticationinfo.T{
		UserID:          n.UserID,
		AuthenticatedAt: deps.Clock.NowUTC(),
		AMR:             amr,
		IdentitySpecs:   identitySpecs,
	}

	// The user has now been reauthenticated. Only when the reauthenticated user
	// is the one named by id_token_hint may the issued tokens continue the hinted
	// session, so record that sid as trusted here -- and only here. Other flows
	// (signup, promote, plain login) have no NodeDidReauthenticate, so they never
	// set it. This explicit check is the validity gate immediately before the set.
	if authflow.GetUserIDHint(ctx) == n.UserID {
		authnInfo.IDTokenHintSID = authflow.GetIDTokenHintSID(ctx)
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
			return deps.AuthenticationInfos.Save(ctx, n.AuthenticationInfoEntry)
		}),
		authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			s := session.GetSession(ctx)
			if idp, ok := s.(*idpsession.IDPSession); ok && idp.GetUserID() == n.UserID {
				err = deps.IDPSessions.Reauthenticate(ctx, idp.ID, n.AuthenticationInfoEntry.T.AMR)
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
