package interaction

import (
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IdentityService interface {
	Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error)
	GetBySpec(spec *identity.Spec) (*identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, spec *identity.Spec) (*identity.Info, error)
	UpdateWithSpec(is *identity.Info, spec *identity.Spec) (*identity.Info, error)
	Create(is *identity.Info) error
	Update(info *identity.Info) error
	Delete(is *identity.Info) error
	CheckDuplicated(info *identity.Info) (*identity.Info, error)
}

type AuthenticatorService interface {
	Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error)
	List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error)
	New(spec *authenticator.Spec, secret string) (*authenticator.Info, error)
	WithSecret(authenticatorInfo *authenticator.Info, secret string) (changed bool, info *authenticator.Info, err error)
	Create(authenticatorInfo *authenticator.Info) error
	Update(authenticatorInfo *authenticator.Info) error
	Delete(authenticatorInfo *authenticator.Info) error
	VerifySecret(info *authenticator.Info, state map[string]string, secret string) error
}

type OOBAuthenticatorProvider interface {
	GenerateCode(secret string, channel authn.AuthenticatorOOBChannel) string
}

type OOBCodeSender interface {
	SendCode(
		channel authn.AuthenticatorOOBChannel,
		target string,
		code string,
		messageType otp.MessageType,
	) (*otp.CodeSendResult, error)
}

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (*anonymous.Request, error)
	GetByKeyID(keyID string) (*anonymous.Identity, error)
	ParseRequest(requestJWT string, identity *anonymous.Identity) (*anonymous.Request, error)
}

type ChallengeProvider interface {
	Consume(token string) (*challenge.Purpose, error)
}

type MFAService interface {
	GenerateDeviceToken() string
	CreateDeviceToken(userID string, token string) (*mfa.DeviceToken, error)
	VerifyDeviceToken(userID string, token string) error
	InvalidateAllDeviceTokens(userID string) error

	GetRecoveryCode(userID string, code string) (*mfa.RecoveryCode, error)
	ConsumeRecoveryCode(rc *mfa.RecoveryCode) error
	GenerateRecoveryCodes() []string
	ReplaceRecoveryCodes(userID string, codes []string) ([]*mfa.RecoveryCode, error)
	ListRecoveryCodes(userID string) ([]*mfa.RecoveryCode, error)
}

type UserService interface {
	Get(id string) (*model.User, error)
	GetRaw(id string) (*user.User, error)
	Create(userID string) (*user.User, error)
	AfterCreate(user *user.User, identities []*identity.Info) error
	UpdateLoginTime(user *model.User, lastLoginAt time.Time) error
}

type HookProvider interface {
	DispatchEvent(payload event.Payload) error
}

type SessionProvider interface {
	MakeSession(*session.Attrs) (*idpsession.IDPSession, string)
	Create(*idpsession.IDPSession) error
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

type ForgotPasswordService interface {
	SendCode(loginID string) error
}

type ResetPasswordService interface {
	ResetPassword(userID string, newPassword string) (oldInfo *authenticator.Info, newInfo *authenticator.Info, err error)
	ResetPasswordByCode(code string, newPassword string) (oldInfo *authenticator.Info, newInfo *authenticator.Info, err error)
	HashCode(code string) (codeHash string)
	AfterResetPasswordByCode(codeHash string) error
}

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) loginid.Normalizer
}

type VerificationService interface {
	GetIdentityVerificationStatus(i *identity.Info) ([]verification.ClaimStatus, error)
	CreateNewCode(id string, info *identity.Info) (*verification.Code, error)
	GetCode(id string) (*verification.Code, error)
	VerifyCode(id string, code string) (*verification.Code, error)
	NewVerifiedClaim(userID string, claimName string, claimValue string) *verification.Claim
	MarkClaimVerified(claim *verification.Claim) error
}

type VerificationCodeSender interface {
	SendCode(code *verification.Code, webStateID string) (*otp.CodeSendResult, error)
}

type CookieFactory interface {
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type Context struct {
	IsDryRun     bool   `wire:"-"`
	IsCommitting bool   `wire:"-"`
	WebStateID   string `wire:"-"`

	Database db.SQLExecutor
	Clock    clock.Clock
	Config   *config.AppConfig

	Identities               IdentityService
	Authenticators           AuthenticatorService
	AnonymousIdentities      AnonymousIdentityProvider
	OOBAuthenticators        OOBAuthenticatorProvider
	OOBCodeSender            OOBCodeSender
	OAuthProviderFactory     OAuthProviderFactory
	MFA                      MFAService
	ForgotPassword           ForgotPasswordService
	ResetPassword            ResetPasswordService
	LoginIDNormalizerFactory LoginIDNormalizerFactory
	Verification             VerificationService
	VerificationCodeSender   VerificationCodeSender

	Challenges           ChallengeProvider
	Users                UserService
	Hooks                HookProvider
	CookieFactory        CookieFactory
	Sessions             SessionProvider
	SessionCookie        idpsession.CookieDef
	MFADeviceTokenCookie mfa.CookieDef
}

var interactionGraphSavePoint savePoint = "interaction_graph"

func (c *Context) initialize() (*Context, error) {
	ctx := *c
	_, err := ctx.Database.ExecWith(interactionGraphSavePoint.New())
	return &ctx, err
}

func (c *Context) commit() error {
	_, err := c.Database.ExecWith(interactionGraphSavePoint.Release())
	return err
}

func (c *Context) rollback() error {
	_, err := c.Database.ExecWith(interactionGraphSavePoint.Rollback())
	return err
}

func (c *Context) perform(effect Effect) error {
	return effect.apply(c)
}
