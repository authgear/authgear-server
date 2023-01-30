package webapp

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type sessionContextKey struct{}

func GetSession(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionContextKey{}).(*Session)
	return s
}

func WithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, session)
}

type SessionOptions struct {
	RedirectURI                string
	KeepAfterFinish            bool
	Prompt                     []string
	Extra                      map[string]interface{}
	Page                       string
	UserIDHint                 string
	WebhookState               string
	UpdatedAt                  time.Time
	CanUseIntentReauthenticate bool
	SuppressIDPSessionCookie   bool
	OAuthProviderAlias         string
	FromAuthzEndpoint          bool
	LoginHint                  string
}

func NewSessionOptionsFromSession(s *Session) SessionOptions {
	return SessionOptions{
		RedirectURI:                s.RedirectURI,
		KeepAfterFinish:            s.KeepAfterFinish,
		Prompt:                     s.Prompt,
		Extra:                      nil, // Omit extra by default
		Page:                       s.Page,
		WebhookState:               s.WebhookState,
		UserIDHint:                 s.UserIDHint,
		CanUseIntentReauthenticate: s.CanUseIntentReauthenticate,
		SuppressIDPSessionCookie:   s.SuppressIDPSessionCookie,
		OAuthProviderAlias:         s.OAuthProviderAlias,
		FromAuthzEndpoint:          s.FromAuthzEndpoint,
		LoginHint:                  s.LoginHint,
	}
}

type Session struct {
	ID string `json:"id"`

	// Steps is a history stack of steps taken within this session.
	Steps []SessionStep `json:"steps,omitempty"`

	// RedirectURI is the URI to redirect to after the completion of session.
	RedirectURI string `json:"redirect_uri,omitempty"`

	// KeepAfterFinish indicates the session would not be deleted after the
	// completion of interaction graph.
	KeepAfterFinish bool `json:"keep_after_finish,omitempty"`

	// Extra is used to store extra information for use of webapp.
	Extra map[string]interface{} `json:"extra"`

	// Prompt is used to indicate requested authentication behavior
	// which includes both supported and unsupported prompt
	Prompt []string `json:"prompt_list,omitempty"`

	// Page is used to indicate the preferred page to show.
	Page string `json:"page,omitempty"`

	// UpdatedAt indicate the session last updated time
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// WebhookState is the state parameter that will pass to the webhook
	// during the web session
	WebhookState string `json:"webhook_state,omitempty"`

	// UserIDHint is the intended user ID.
	// It is expected that the authenticated user is indicated by this user ID,
	// otherwise it is an error.
	UserIDHint string `json:"user_id_hint,omitempty"`

	// CanUseIntentReauthenticate indicates whether IntentReauthenticate can be used.
	CanUseIntentReauthenticate bool `json:"can_use_intent_reauthenticate,omitempty"`

	// SuppressIDPSessionCookie indicates whether IDP session cookie should neither be read or written.
	SuppressIDPSessionCookie bool `json:"suppress_idp_session_cookie,omitempty"`

	// OAuthProviderAlias is used to auto redirect user to the given oauth provider in the login page
	OAuthProviderAlias string `json:"oauth_provider_alias,omitempty"`

	// FromAuthzEndpoint indicates whether the web session is created from the authorization endpoint
	FromAuthzEndpoint bool `json:"from_authz_endpoint,omitempty"`

	// LoginHint is the OIDC login_hint parameter.
	LoginHint string `json:"login_hint,omitempty"`
}

func newSessionID() string {
	const (
		idAlphabet string = base32.Alphabet
		idLength   int    = 32
	)
	return corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand)
}

func NewSession(options SessionOptions) *Session {
	s := &Session{
		ID:                         newSessionID(),
		RedirectURI:                options.RedirectURI,
		KeepAfterFinish:            options.KeepAfterFinish,
		Extra:                      make(map[string]interface{}),
		Prompt:                     options.Prompt,
		Page:                       options.Page,
		UpdatedAt:                  options.UpdatedAt,
		WebhookState:               options.WebhookState,
		UserIDHint:                 options.UserIDHint,
		CanUseIntentReauthenticate: options.CanUseIntentReauthenticate,
		SuppressIDPSessionCookie:   options.SuppressIDPSessionCookie,
		OAuthProviderAlias:         options.OAuthProviderAlias,
		FromAuthzEndpoint:          options.FromAuthzEndpoint,
		LoginHint:                  options.LoginHint,
	}
	for k, v := range options.Extra {
		s.Extra[k] = v
	}
	return s
}

func (s *Session) CurrentStep() SessionStep {
	return s.Steps[len(s.Steps)-1]
}
