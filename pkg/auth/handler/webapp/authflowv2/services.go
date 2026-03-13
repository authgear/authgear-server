package authflowv2

import (
	"context"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
)

type SettingsIdentityService interface {
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	ListCandidates(ctx context.Context, userID string) ([]identity.Candidate, error)
	GetWithUserID(ctx context.Context, userID string, identityID string) (*identity.Info, error)
	GetBySpecWithUserID(ctx context.Context, userID string, spec *identity.Spec) (*identity.Info, error)
}

type SettingsEndpointsProvider interface {
	SSOCallbackURL(alias string) *url.URL
	SharedSSOCallbackURL() *url.URL
}

type SettingsOAuthStateStore interface {
	GenerateState(ctx context.Context, state *webappoauth.WebappOAuthState) (stateToken string, err error)
}

type SettingsAuthenticatorService interface {
	List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
}

type SettingsVerificationService interface {
	GetVerificationStatuses(ctx context.Context, is []*identity.Info) (map[string][]verification.ClaimStatus, error)
}

type SettingsSessionManager interface {
	List(ctx context.Context, userID string) ([]session.ListableSession, error)
	Get(ctx context.Context, id string) (session.ListableSession, error)
	RevokeWithEvent(ctx context.Context, s session.SessionBase, isTermination bool, isAdminAPI bool) error
	TerminateAllExcept(ctx context.Context, userID string, currentSession session.ResolvedSession, isAdminAPI bool) error
}

type SettingsAuthorizationService interface {
	GetByID(ctx context.Context, id string) (*oauth.Authorization, error)
	ListByUser(ctx context.Context, userID string, filters ...oauth.AuthorizationFilter) ([]*oauth.Authorization, error)
	Delete(ctx context.Context, a *oauth.Authorization) error
}

type SettingsSessionListingService interface {
	FilterForDisplay(ctx context.Context, sessions []session.ListableSession, currentSession session.ResolvedSession) ([]*sessionlisting.Session, error)
}

type SettingsMFAService interface {
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(ctx context.Context, userID string) error
}
