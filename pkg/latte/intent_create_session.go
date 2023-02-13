package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentCreateSession{})
}

var IntentCreateSessionSchema = validation.NewSimpleSchema(`{}`)

type IntentCreateSession struct {
	UserID       string               `json:"user_id"`
	CreateReason session.CreateReason `json:"create_reason"`
	AMR          []string             `json:"amr,omitempty"`
	SkipCreate   bool                 `json:"skip_create"`
}

func (*IntentCreateSession) Kind() string {
	return "latte.IntentCreateSession"
}

func (*IntentCreateSession) JSONSchema() *validation.SimpleSchema {
	return IntentCreateSessionSchema
}

func (*IntentCreateSession) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentCreateSession) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	attrs := session.NewAttrs(i.UserID)
	attrs.SetAMR(i.AMR)
	s, token := deps.IDPSessions.MakeSession(attrs)
	sessionCookie := deps.Cookies.ValueCookie(deps.SessionCookie.Def, token)

	authnInfo := s.GetAuthenticationInfo()
	authnInfoEntry := authenticationinfo.NewEntry(authnInfo)
	authnInfoCookie := deps.Cookies.ValueCookie(
		authenticationinfo.CookieDef,
		authnInfoEntry.ID,
	)

	sameSiteStrictCookie := deps.Cookies.ValueCookie(
		deps.SessionCookie.SameSiteStrictDef,
		"true",
	)

	if i.SkipCreate {
		s = nil
		sessionCookie = nil
	}

	return workflow.NewNodeSimple(&NodeDoCreateSession{
		UserID:                   i.UserID,
		CreateReason:             i.CreateReason,
		Session:                  s,
		AuthenticationInfoEntry:  authnInfoEntry,
		SessionCookie:            sessionCookie,
		SameSiteStrictCookie:     sameSiteStrictCookie,
		AuthenticationInfoCookie: authnInfoCookie,
	}), nil
}

func (*IntentCreateSession) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentCreateSession) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
