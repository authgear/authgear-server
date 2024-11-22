package workflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type Session struct {
	WorkflowID string `json:"workflow_id"`

	OAuthSessionID           string `json:"oauth_session_id,omitempty"`
	ClientID                 string `json:"client_id,omitempty"`
	RedirectURI              string `json:"redirect_uri,omitempty"`
	SuppressIDPSessionCookie bool   `json:"suppress_idp_session_cookie,omitempty"`
	State                    string `json:"state,omitempty"`
	XState                   string `json:"x_state,omitempty"`
	UILocales                string `json:"ui_locales,omitempty"`
	UserAgentID              string `json:"user_agent_id,omitempty"`
	// UserIDHint is for reauthentication.
	UserIDHint string `json:"user_id_hint,omitempty"`
}

type SessionOutput struct {
	WorkflowID  string `json:"workflow_id"`
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type SessionOptions struct {
	OAuthSessionID           string
	ClientID                 string
	RedirectURI              string
	SuppressIDPSessionCookie bool
	State                    string
	XState                   string
	UILocales                string
	UserAgentID              string
	// UserIDHint is for reauthentication.
	UserIDHint string
}

func (s *SessionOptions) PartiallyMergeFrom(o *SessionOptions) *SessionOptions {
	out := &SessionOptions{}
	if s != nil {
		out.OAuthSessionID = s.OAuthSessionID
		out.ClientID = s.ClientID
		out.RedirectURI = s.RedirectURI
		out.SuppressIDPSessionCookie = s.SuppressIDPSessionCookie
		out.State = s.State
		out.XState = s.XState
		out.UILocales = s.UILocales
		out.UserIDHint = s.UserIDHint
	}
	if o != nil {
		if o.ClientID != "" {
			out.ClientID = o.ClientID
		}
		if o.State != "" {
			out.State = o.State
		}
		if o.XState != "" {
			out.XState = o.XState
		}
		if o.UILocales != "" {
			out.UILocales = o.UILocales
		}
		if o.UserIDHint != "" {
			out.UserIDHint = o.UserIDHint
		}
	}
	return out
}

func NewSession(opts *SessionOptions) *Session {
	return &Session{
		WorkflowID:               newWorkflowID(),
		OAuthSessionID:           opts.OAuthSessionID,
		ClientID:                 opts.ClientID,
		RedirectURI:              opts.RedirectURI,
		SuppressIDPSessionCookie: opts.SuppressIDPSessionCookie,
		State:                    opts.State,
		XState:                   opts.XState,
		UILocales:                opts.UILocales,
		UserAgentID:              opts.UserAgentID,
		UserIDHint:               opts.UserIDHint,
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
	ctx = context.WithValue(ctx, contextKeyOAuthSessionID, s.OAuthSessionID)

	if s.ClientID != "" {
		key := otelauthgear.AttributeKeyClientID
		val := key.String(s.ClientID)
		ctx = context.WithValue(ctx, key, val)
	}

	ctx = uiparam.WithUIParam(ctx, &uiparam.T{
		ClientID:  s.ClientID,
		UILocales: s.UILocales,
		State:     s.State,
		XState:    s.XState,
	})
	ctx = intl.WithPreferredLanguageTags(ctx, intl.ParseUILocales(s.UILocales))
	ctx = context.WithValue(ctx, contextKeySuppressIDPSessionCookie, s.SuppressIDPSessionCookie)
	ctx = context.WithValue(ctx, contextKeyWorkflowID, s.WorkflowID)
	ctx = context.WithValue(ctx, contextKeyUserIDHint, s.UserIDHint)
	return ctx
}
