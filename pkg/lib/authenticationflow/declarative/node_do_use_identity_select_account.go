package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func init() {
	authflow.RegisterNode(&NodeDoUseIdentitySelectAccount{})
}

type NodeDoUseIdentitySelectAccount struct {
	UserID      string `json:"user_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	SessionType string `json:"session_type,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoUseIdentitySelectAccount{}
var _ authflow.Milestone = &NodeDoUseIdentitySelectAccount{}
var _ MilestoneDoUseUser = &NodeDoUseIdentitySelectAccount{}
var _ MilestoneDoUseExistingSession = &NodeDoUseIdentitySelectAccount{}
var _ authflow.InputReactor = &NodeDoUseIdentitySelectAccount{}

func (n *NodeDoUseIdentitySelectAccount) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return nil, nil
}

func (n *NodeDoUseIdentitySelectAccount) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return NewNodePostIdentified(ctx, deps, flows, &NodePostIdentifiedOptions{
		Identification: model.Identification{
			Identification: model.AuthenticationFlowIdentificationSelectAccount,
			Identity:       nil,
			IDToken:        nil,
		},
	})
}

func NewNodeDoUseIdentitySelectAccount(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, expectedUserID string) (authflow.ReactToResult, error) {
	userID, sessionID, sessionType, err := resolveSelectAccountSession(ctx, expectedUserID)
	if err != nil {
		return nil, err
	}

	n := &NodeDoUseIdentitySelectAccount{UserID: userID, SessionID: sessionID, SessionType: string(sessionType)}
	return authflow.NewNodeSimple(n), nil
}

func (*NodeDoUseIdentitySelectAccount) Kind() string {
	return "NodeDoUseIdentitySelectAccount"
}

func (*NodeDoUseIdentitySelectAccount) Milestone() {}

func (n *NodeDoUseIdentitySelectAccount) MilestoneDoUseUser() string {
	return n.UserID
}

// MilestoneDoUseExistingSession signals that this session must be reused
// as-is by the flow's final NodeDoCreateSession, never rotated or renewed —
// per docs/specs/custom-ui-select-account.md's "Completing identification
// with the existing session".
func (n *NodeDoUseIdentitySelectAccount) MilestoneDoUseExistingSession() (sessionType session.Type, sessionID string) {
	return session.Type(n.SessionType), n.SessionID
}

// resolveSelectAccountSession checks that the session cookie still resolves
// to the same user recorded when the select_account option was constructed
// (see NewIdentificationOptionsSelectAccount), and returns that session's
// own ID/type so the caller can record it for later reuse (see
// MilestoneDoUseExistingSession). Used both when select_account completes
// directly in a login flow, and when signup_login is about to switch into
// the target login_flow (see IntentLookupIdentitySelectAccount) — the
// latter only needs the userID for its own re-check, the session's
// ID/type are re-resolved independently once the target login_flow
// actually completes.
func resolveSelectAccountSession(ctx context.Context, expectedUserID string) (userID string, sessionID string, sessionType session.Type, err error) {
	sess := session.GetSession(ctx)
	if sess == nil {
		return "", "", "", ErrSelectAccountSessionChanged
	}
	userID = sess.GetAuthenticationInfo().UserID
	if userID != expectedUserID {
		return "", "", "", ErrSelectAccountSessionChanged
	}
	return userID, sess.SessionID(), sess.SessionType(), nil
}
