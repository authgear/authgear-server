package authenticationflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type Session struct {
	FlowID         string `json:"flow_id"`
	UserAgentID    string `json:"user_agent_id,omitempty"`
	OAuthSessionID string `json:"oauth_session_id,omitempty"`

	ClientID    string   `json:"client_id,omitempty"`
	RedirectURI string   `json:"redirect_uri,omitempty"`
	Prompt      []string `json:"prompt,omitempty"`
	State       string   `json:"state,omitempty"`
	XState      string   `json:"x_state,omitempty"`
	UILocales   string   `json:"ui_locales,omitempty"`

	SuppressIDPSessionCookie bool `json:"suppress_idp_session_cookie,omitempty"`
}

type SessionOutput struct {
	FlowID      string `json:"flow_id"`
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type SessionOptions struct {
	UserAgentID    string
	OAuthSessionID string

	ClientID    string
	RedirectURI string
	Prompt      []string
	State       string
	XState      string
	UILocales   string

	SuppressIDPSessionCookie bool
}

func (s *SessionOptions) PartiallyMergeFrom(o *SessionOptions) *SessionOptions {
	out := &SessionOptions{}
	if s != nil {
		out.OAuthSessionID = s.OAuthSessionID

		out.ClientID = s.ClientID
		out.RedirectURI = s.RedirectURI
		out.Prompt = s.Prompt
		out.State = s.State
		out.XState = s.XState
		out.UILocales = s.UILocales

		out.SuppressIDPSessionCookie = s.SuppressIDPSessionCookie
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
		FlowID:         newFlowID(),
		UserAgentID:    opts.UserAgentID,
		OAuthSessionID: opts.OAuthSessionID,

		ClientID:    opts.ClientID,
		RedirectURI: opts.RedirectURI,
		Prompt:      opts.Prompt,
		State:       opts.State,
		UILocales:   opts.UILocales,

		SuppressIDPSessionCookie: opts.SuppressIDPSessionCookie,
		XState:                   opts.XState,
	}
}

func (s *Session) ToOutput() *SessionOutput {
	return &SessionOutput{
		FlowID:      s.FlowID,
		ClientID:    s.ClientID,
		RedirectURI: s.RedirectURI,
	}
}

func (s *Session) MakeContext(ctx context.Context, deps *Dependencies, publicFlow PublicFlow) (context.Context, error) {
	ctx = context.WithValue(ctx, contextKeyOAuthSessionID, s.OAuthSessionID)

	ctx = uiparam.WithUIParam(ctx, &uiparam.T{
		ClientID:  s.ClientID,
		Prompt:    s.Prompt,
		UILocales: s.UILocales,
		State:     s.State,
		XState:    s.XState,
	})

	ctx = intl.WithPreferredLanguageTags(ctx, intl.ParseUILocales(s.UILocales))

	ctx = context.WithValue(ctx, contextKeySuppressIDPSessionCookie, s.SuppressIDPSessionCookie)

	ctx = context.WithValue(ctx, contextKeyFlowID, s.FlowID)

	flowReference := publicFlow.FlowFlowReference()
	ctx = context.WithValue(ctx, contextKeyFlowReference, flowReference)

	flowRootObject, err := publicFlow.FlowRootObject(deps)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, contextKeyFlowRootObject, flowRootObject)

	return ctx, nil
}
