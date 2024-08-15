package authenticationflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

// Session must not contain web session ID.
// This is to ensure webapp does not have privilege in authflow.
type Session struct {
	FlowID string `json:"flow_id"`

	OAuthSessionID string `json:"oauth_session_id,omitempty"`
	SAMLSessionID  string `json:"saml_session_id,omitempty"`

	ClientID    string   `json:"client_id,omitempty"`
	RedirectURI string   `json:"redirect_uri,omitempty"`
	Prompt      []string `json:"prompt,omitempty"`
	State       string   `json:"state,omitempty"`
	XState      string   `json:"x_state,omitempty"`
	UILocales   string   `json:"ui_locales,omitempty"`

	BotProtectionVerificationResult *BotProtectionVerificationResult `json:"bot_protection_verification_result,omitempty"`
	IDToken                         string                           `json:"id_token,omitempty"`
	SuppressIDPSessionCookie        bool                             `json:"suppress_idp_session_cookie,omitempty"`
	UserIDHint                      string                           `json:"user_id_hint,omitempty"`
	LoginHint                       string                           `json:"login_hint,omitempty"`
}

type SessionOutput struct {
	FlowID      string `json:"flow_id"`
	ClientID    string `json:"client_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type SessionOptions struct {
	OAuthSessionID string
	SAMLSessionID  string

	ClientID    string
	RedirectURI string
	Prompt      []string
	State       string
	XState      string
	UILocales   string

	BotProtectionVerificationResult *BotProtectionVerificationResult
	IDToken                         string
	SuppressIDPSessionCookie        bool
	UserIDHint                      string
	LoginHint                       string
}

func (s *SessionOptions) PartiallyMergeFrom(o *SessionOptions) *SessionOptions {
	out := &SessionOptions{}
	if s != nil {
		out.OAuthSessionID = s.OAuthSessionID
		out.SAMLSessionID = s.SAMLSessionID

		out.ClientID = s.ClientID
		out.RedirectURI = s.RedirectURI
		out.Prompt = s.Prompt
		out.State = s.State
		out.XState = s.XState
		out.UILocales = s.UILocales

		out.BotProtectionVerificationResult = s.BotProtectionVerificationResult
		out.IDToken = s.IDToken
		out.SuppressIDPSessionCookie = s.SuppressIDPSessionCookie
		out.UserIDHint = s.UserIDHint
		out.LoginHint = s.LoginHint
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
		FlowID: newFlowID(),

		OAuthSessionID: opts.OAuthSessionID,
		SAMLSessionID:  opts.SAMLSessionID,

		ClientID:    opts.ClientID,
		RedirectURI: opts.RedirectURI,
		Prompt:      opts.Prompt,
		State:       opts.State,
		XState:      opts.XState,
		UILocales:   opts.UILocales,

		BotProtectionVerificationResult: opts.BotProtectionVerificationResult,
		IDToken:                         opts.IDToken,
		SuppressIDPSessionCookie:        opts.SuppressIDPSessionCookie,
		UserIDHint:                      opts.UserIDHint,
		LoginHint:                       opts.LoginHint,
	}
}

func (s *Session) ToOutput() *SessionOutput {
	return &SessionOutput{
		FlowID:      s.FlowID,
		ClientID:    s.ClientID,
		RedirectURI: s.RedirectURI,
	}
}

func (s *Session) MakeContext(ctx context.Context, deps *Dependencies) (context.Context, error) {
	ctx = context.WithValue(ctx, contextKeyOAuthSessionID, s.OAuthSessionID)
	ctx = context.WithValue(ctx, contextKeySAMLSessionID, s.SAMLSessionID)

	ctx = uiparam.WithUIParam(ctx, &uiparam.T{
		ClientID:  s.ClientID,
		Prompt:    s.Prompt,
		UILocales: s.UILocales,
		State:     s.State,
		XState:    s.XState,
	})

	if s.UILocales != "" {
		tags := intl.ParseUILocales(s.UILocales)
		ctx = intl.WithPreferredLanguageTags(ctx, tags)
	} else {
		acceptLanguage := deps.HTTPRequest.Header.Get("Accept-Language")
		tags := intl.ParseAcceptLanguage(acceptLanguage)
		ctx = intl.WithPreferredLanguageTags(ctx, tags)
	}

	ctx = context.WithValue(ctx, contextKeyBotProtectionVerificationResult, s.BotProtectionVerificationResult)
	ctx = context.WithValue(ctx, contextKeyIDToken, s.IDToken)
	ctx = context.WithValue(ctx, contextKeySuppressIDPSessionCookie, s.SuppressIDPSessionCookie)
	ctx = context.WithValue(ctx, contextKeyUserIDHint, s.UserIDHint)
	ctx = context.WithValue(ctx, contextKeyLoginHint, s.LoginHint)

	ctx = context.WithValue(ctx, contextKeyFlowID, s.FlowID)

	return ctx, nil
}

func (s *Session) SetBotProtectionVerificationResult(result *BotProtectionVerificationResult) {
	s.BotProtectionVerificationResult = result
}
