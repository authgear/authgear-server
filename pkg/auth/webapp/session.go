package webapp

import (
	"context"

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
	KeepAfterFinish bool
	UILocales       string
}

type SessionStep struct {
	GraphID string `json:"graph_id"`
	Path    string `json:"path"`
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

	// UILocales are the locale to be used to render UI, passed in from OAuth
	// flow or query parameter.
	UILocales string `json:"ui_locales,omitempty"`
}

func NewSession(options SessionOptions) *Session {
	const (
		idAlphabet string = base32.Alphabet
		idLength   int    = 32
	)
	return &Session{
		ID:              corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand),
		RedirectURI:     options.RedirectURI,
		KeepAfterFinish: options.KeepAfterFinish,
		Extra:           make(map[string]interface{}),
		UILocales:       options.UILocales,
	}
}
