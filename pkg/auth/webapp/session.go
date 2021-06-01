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
	RedirectURI     string
	ClientID        string
	KeepAfterFinish bool
	Prompt          []string
	Extra           map[string]interface{}
	WebhookState    string
	UpdatedAt       time.Time
}

func NewSessionOptionsFromSession(s *Session) SessionOptions {
	return SessionOptions{
		RedirectURI:     s.RedirectURI,
		ClientID:        s.ClientID,
		KeepAfterFinish: s.KeepAfterFinish,
		Prompt:          s.Prompt,
		Extra:           nil, // Omit extra by default
		WebhookState:    s.WebhookState,
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

	// ClientID is the client ID associated with this session.
	ClientID string `json:"client_id,omitempty"`

	// Extra is used to store extra information for use of webapp.
	Extra map[string]interface{} `json:"extra"`

	// Prompt is used to indicate requested authentication behavior
	// which includes both supported and unsupported prompt
	Prompt []string `json:"prompt_list,omitempty"`

	// FIXME(webapp): remove LegacyPrompt in the next deployment
	// LegacyPrompt is legacy prompt parameter
	// which should restore to prompt list after unmarshal
	LegacyPrompt string `json:"prompt,omitempty"`

	// UpdatedAt indicate the session last updated time
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// WebhookState is the state parameter that will pass to the webhook
	// during the web session
	WebhookState string `json:"webhook_state,omitempty"`
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
		ID:              newSessionID(),
		RedirectURI:     options.RedirectURI,
		KeepAfterFinish: options.KeepAfterFinish,
		ClientID:        options.ClientID,
		Extra:           make(map[string]interface{}),
		Prompt:          options.Prompt,
		UpdatedAt:       options.UpdatedAt,
		WebhookState:    options.WebhookState,
	}
	for k, v := range options.Extra {
		s.Extra[k] = v
	}
	if s.RedirectURI == "" {
		s.RedirectURI = "/"
	}
	return s
}

func (s *Session) CurrentStep() SessionStep {
	return s.Steps[len(s.Steps)-1]
}
