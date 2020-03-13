package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// Provider handles authentication process.
type Provider interface {
	// BeginSession creates a new authentication session.
	BeginSession(client config.OAuthClientConfiguration, userID string, prin principal.Principal, reason session.CreateReason) (*Session, error)

	// StepSession update current step of an authentication session and return authentication result.
	StepSession(s *Session) (Result, error)

	// MakeResult loads related data for an existing session to create authentication result.
	MakeResult(client config.OAuthClientConfiguration, s *session.Session) (Result, error)

	// Resolve resolves token to authentication session.
	ResolveSession(token string) (*Session, error)
}
