package authenticationflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type Session struct {
	FlowID string `json:"flow_id"`

	ClientID                 string `json:"client_id,omitempty"`
	RedirectURI              string `json:"redirect_uri,omitempty"`
	SuppressIDPSessionCookie bool   `json:"suppress_idp_session_cookie,omitempty"`
	State                    string `json:"state,omitempty"`
	XState                   string `json:"x_state,omitempty"`
	UILocales                string `json:"ui_locales,omitempty"`
	UserAgentID              string `json:"user_agent_id,omitempty"`
}

type SessionOutput struct {
	FlowID      string `json:"flow_id"`
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type SessionOptions struct {
	ClientID                 string
	RedirectURI              string
	SuppressIDPSessionCookie bool
	State                    string
	XState                   string
	UILocales                string
	UserAgentID              string
}

func (s *SessionOptions) PartiallyMergeFrom(o *SessionOptions) *SessionOptions {
	out := &SessionOptions{}
	if s != nil {
		out.ClientID = s.ClientID
		out.RedirectURI = s.RedirectURI
		out.SuppressIDPSessionCookie = s.SuppressIDPSessionCookie
		out.State = s.State
		out.XState = s.XState
		out.UILocales = s.UILocales
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
	}
	return out
}

func NewSession(opts *SessionOptions) *Session {
	return &Session{
		FlowID:                   newFlowID(),
		ClientID:                 opts.ClientID,
		RedirectURI:              opts.RedirectURI,
		SuppressIDPSessionCookie: opts.SuppressIDPSessionCookie,
		State:                    opts.State,
		XState:                   opts.XState,
		UILocales:                opts.UILocales,
		UserAgentID:              opts.UserAgentID,
	}
}

func (s *Session) ToOutput() *SessionOutput {
	return &SessionOutput{
		FlowID:      s.FlowID,
		ClientID:    s.ClientID,
		RedirectURI: s.RedirectURI,
	}
}

func (s *Session) Context(ctx context.Context) context.Context {
	ctx = uiparam.WithUIParam(ctx, &uiparam.T{
		ClientID:  s.ClientID,
		UILocales: s.UILocales,
		State:     s.State,
		XState:    s.XState,
	})
	ctx = intl.WithPreferredLanguageTags(ctx, intl.ParseUILocales(s.UILocales))
	ctx = context.WithValue(ctx, contextKeySuppressIDPSessionCookie, s.SuppressIDPSessionCookie)
	ctx = context.WithValue(ctx, contextKeyFlowID, s.FlowID)
	return ctx
}
