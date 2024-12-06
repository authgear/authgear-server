package webapp

import (
	"context"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/sessionlisting"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// This file contains the interfaces that are used by v2 code.
// Ideally we should remove these interfaces by using local-to-handler interfaces.
// So please do not add more interfaces here!

type ErrorService interface {
	HasError(ctx context.Context, r *http.Request) bool
}

type FlashMessage interface {
	Flash(rw http.ResponseWriter, messageType string)
}

type MeterService interface {
	TrackPageView(ctx context.Context, visitorID string, pageType meter.PageType) error
}

type TutorialCookie interface {
	Pop(r *http.Request, rw http.ResponseWriter, name httputil.TutorialCookieName) bool
}

type SettingsDeleteAccountUserService interface {
	ScheduleDeletionByEndUser(ctx context.Context, userID string) error
}

type SettingsDeleteAccountOAuthSessionService interface {
	Get(ctx context.Context, entryID string) (*oauthsession.Entry, error)
	Save(ctx context.Context, entry *oauthsession.Entry) error
}

type SettingsDeleteAccountSessionStore interface {
	Create(ctx context.Context, session *webapp.Session) (err error)
	Delete(ctx context.Context, id string) (err error)
	Update(ctx context.Context, session *webapp.Session) (err error)
}

type SettingsDeleteAccountAuthenticationInfoService interface {
	Save(ctx context.Context, entry *authenticationinfo.Entry) (err error)
}

type SettingsDeleteAccountSuccessUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type SettingsDeleteAccountSuccessAuthenticationInfoService interface {
	Get(ctx context.Context, entryID string) (entry *authenticationinfo.Entry, err error)
}

type SettingsProfileEditUserService interface {
	GetRaw(ctx context.Context, id string) (*user.User, error)
}

type SettingsProfileEditStdAttrsService interface {
	UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type SettingsProfileEditCustomAttrsService interface {
	UpdateCustomAttributesWithForm(ctx context.Context, role accesscontrol.Role, userID string, jsonPointerMap map[string]string) error
}

type SettingsVerificationService interface {
	GetVerificationStatuses(ctx context.Context, is []*identity.Info) (map[string][]verification.ClaimStatus, error)
}

type SettingsEndpointsProvider interface {
	SSOCallbackURL(provider string) *url.URL
}

type SettingsOAuthStateStore interface {
	GenerateState(ctx context.Context, state *webappoauth.WebappOAuthState) (stateToken string, err error)
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

type SettingsIdentityService interface {
	ListByUser(ctx context.Context, userID string) ([]*identity.Info, error)
	ListCandidates(ctx context.Context, userID string) ([]identity.Candidate, error)
}

type SettingsMFAService interface {
	ListRecoveryCodes(ctx context.Context, userID string) ([]*mfa.RecoveryCode, error)
	InvalidateAllDeviceTokens(ctx context.Context, userID string) error
}

type AuthflowSignupEndpointsProvider interface {
	SSOCallbackURL(alias string) *url.URL
}

type PasswordPolicy interface {
	PasswordPolicy() []password.Policy
	PasswordRules() string
}
