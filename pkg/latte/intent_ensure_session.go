package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentEnsureSession{})
}

var IntentEnsureSessionSchema = validation.NewSimpleSchema(`{}`)

type EnsureSessionMode string

const (
	EnsureSessionModeDefault        EnsureSessionMode = ""
	EnsureSessionModeCreate         EnsureSessionMode = "create"
	EnsureSessionModeUpdateOrCreate EnsureSessionMode = "update_or_create"
	EnsureSessionModeNoop           EnsureSessionMode = "noop"
)

type IntentEnsureSession struct {
	UserID       string               `json:"user_id"`
	CreateReason session.CreateReason `json:"create_reason"`
	AMR          []string             `json:"amr,omitempty"`
	Mode         EnsureSessionMode    `json:"mode"`
}

func (*IntentEnsureSession) Kind() string {
	return "latte.IntentEnsureSession"
}

func (*IntentEnsureSession) JSONSchema() *validation.SimpleSchema {
	return IntentEnsureSessionSchema
}

func (*IntentEnsureSession) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentEnsureSession) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	attrs := session.NewAttrs(i.UserID)
	attrs.SetAMR(i.AMR)
	var sessionToCreate *idpsession.IDPSession = nil
	newSession, token := deps.IDPSessions.MakeSession(attrs)
	sessionToCreate = newSession
	sessionCookie := deps.Cookies.ValueCookie(deps.SessionCookie.Def, token)

	mode := i.Mode
	if mode == EnsureSessionModeDefault {
		mode = EnsureSessionModeCreate
	}

	sameSiteStrictCookie := deps.Cookies.ValueCookie(
		deps.SessionCookie.SameSiteStrictDef,
		"true",
	)

	var updateSessionID string
	var updateSessionAMR []string
	if mode == EnsureSessionModeUpdateOrCreate {
		s := session.GetSession(ctx)
		if idp, ok := s.(*idpsession.IDPSession); ok && idp.GetUserID() == i.UserID {
			updateSessionID = idp.ID
			updateSessionAMR = i.AMR
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

	var authnInfo authenticationinfo.T
	if sessionToCreate == nil {
		authnInfo = newSession.GetAuthenticationInfo()
	} else {
		authnInfo = sessionToCreate.CreateNewAuthenticationInfoByThisSession()
	}
	authnInfo.ShouldFireAuthenticatedEventWhenIssueOfflineGrant = mode == EnsureSessionModeNoop && i.CreateReason == session.CreateReasonLogin
	authnInfoEntry := authenticationinfo.NewEntry(
		authnInfo,
		workflow.GetOAuthSessionID(ctx),
	)

	return workflow.NewNodeSimple(&NodeDoEnsureSession{
		UserID:                  i.UserID,
		CreateReason:            i.CreateReason,
		SessionToCreate:         sessionToCreate,
		AuthenticationInfoEntry: authnInfoEntry,
		SessionCookie:           sessionCookie,
		UpdateSessionID:         updateSessionID,
		UpdateSessionAMR:        updateSessionAMR,
		SameSiteStrictCookie:    sameSiteStrictCookie,
	}), nil
}

func (*IntentEnsureSession) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentEnsureSession) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentEnsureSession) GetSession(w *workflow.Workflow) *idpsession.IDPSession {
	node, ok := workflow.FindSingleNode[*NodeDoEnsureSession](w)
	if !ok {
		return nil
	}

	return node.SessionToCreate
}
