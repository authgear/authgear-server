package workflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/clientid"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
)

type Session struct {
	WorkflowID string `json:"workflow_id"`

	ClientID                 string `json:"client_id,omitempty"`
	RedirectURI              string `json:"redirect_uri,omitempty"`
	SuppressIDPSessionCookie bool   `json:"suppress_idp_session_cookie,omitempty"`
	State                    string `json:"state,omitempty"`
	UILocales                string `json:"ui_locales,omitempty"`
}

type SessionOutput struct {
	WorkflowID  string `json:"workflow_id"`
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type SessionOptions struct {
	ClientID                 string
	RedirectURI              string
	SuppressIDPSessionCookie bool
	State                    string
	UILocales                string
}

func NewSession(opts *SessionOptions) *Session {
	return &Session{
		WorkflowID:               newWorkflowID(),
		ClientID:                 opts.ClientID,
		RedirectURI:              opts.RedirectURI,
		SuppressIDPSessionCookie: opts.SuppressIDPSessionCookie,
		State:                    opts.State,
		UILocales:                opts.UILocales,
	}
}

func (s *Session) ToOutput() *SessionOutput {
	return &SessionOutput{
		WorkflowID:  s.WorkflowID,
		ClientID:    s.ClientID,
		RedirectURI: s.RedirectURI,
	}
}

func (s *Session) Context(ctx context.Context) context.Context {
	ctx = clientid.WithClientID(ctx, s.ClientID)
	ctx = uiparam.WithUIParam(ctx, s.State, s.UILocales)
	ctx = context.WithValue(ctx, contextKeySuppressIDPSessionCookie, s.SuppressIDPSessionCookie)
	ctx = context.WithValue(ctx, contextKeyState, s.State)
	ctx = context.WithValue(ctx, contextKeyWorkflowID, s.WorkflowID)
	return ctx
}
