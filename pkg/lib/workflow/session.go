package workflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/clientid"
)

type Session struct {
	WorkflowID string `json:"workflow_id"`

	ClientID                 string `json:"client_id,omitempty"`
	RedirectURI              string `json:"redirect_uri,omitempty"`
	SuppressIDPSessionCookie bool   `json:"suppress_idp_session_cookie,omitempty"`
	State                    string `json:"state,omitempty"`
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
}

func NewSession(opts *SessionOptions) *Session {
	return &Session{
		WorkflowID:               newWorkflowID(),
		ClientID:                 opts.ClientID,
		RedirectURI:              opts.RedirectURI,
		SuppressIDPSessionCookie: opts.SuppressIDPSessionCookie,
		State:                    opts.State,
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
	ctx = context.WithValue(ctx, contextKeySuppressIDPSessionCookie, s.SuppressIDPSessionCookie)
	ctx = context.WithValue(ctx, contextKeyState, s.State)
	ctx = context.WithValue(ctx, contextKeyWorkflowID, s.WorkflowID)
	return ctx
}
