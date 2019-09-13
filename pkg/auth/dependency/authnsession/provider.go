package authnsession

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

type ResolveUserIDOptions struct {
	MFACase ResolveUserIDMFACase
}

type ResolveUserIDMFACase int

const (
	ResolveUserIDMFACaseAlwaysAccept ResolveUserIDMFACase = iota
	ResolveUserIDMfaCaseOnlyWhenNoAuthenticators
)

// Provider manipulates authentication session
type Provider interface {
	// NewWithToken decodes an authentication session from a token.
	NewWithToken(token string) (*auth.AuthnSession, error)
	// NewFromScratch creates a new authentication session.
	NewFromScratch(userID string, prin principal.Principal, reason event.SessionCreateReason) (*auth.AuthnSession, error)
	// GenerateResponseAndUpdateLastLoginAt generates authentication response and update last_login_at
	// if the response is AuthResponse.
	GenerateResponseAndUpdateLastLoginAt(session auth.AuthnSession) (interface{}, error)
	// WriteResponse alters the response, write Cookies and write HTTP Body. It should be used in a defer block.
	// It should be used in most cases.
	WriteResponse(w http.ResponseWriter, resp interface{}, err error)
	// AlterResponse alters the response and write Cookies.
	// It should only be used when the response is not given in HTTP Body.
	AlterResponse(w http.ResponseWriter, resp interface{}, err error) interface{}

	// ResolveUserID resolves user ID from auth context or the token.
	ResolveUserID(authContext auth.ContextGetter, authnSessionToken string, options ResolveUserIDOptions) (string, error)
}
