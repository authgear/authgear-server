package newinteraction

import (
	"time"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/auth/dependency/user"
	"github.com/authgear/authgear-server/pkg/auth/event"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/otp"
)

type IdentityProvider interface {
	Get(userID string, typ authn.IdentityType, id string) (*identity.Info, error)
	// FIXME: no need to return userID, now identity.Info has it
	// FIXME: take *identity.Spec instead
	GetByClaims(typ authn.IdentityType, claims map[string]interface{}) (string, *identity.Info, error)
	ListByUser(userID string) ([]*identity.Info, error)
	New(userID string, typ authn.IdentityType, claims map[string]interface{}) (*identity.Info, error)
	CreateAll(is []*identity.Info) error
	UpdateAll(is []*identity.Info) error
	Validate(is []*identity.Info) error
}

type AuthenticatorProvider interface {
	// FIXME: cleanup user ID, authenticator.Info & Spec has it
	Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error)
	List(userID string, typ authn.AuthenticatorType) ([]*authenticator.Info, error)
	ListByIdentity(userID string, ii *identity.Info) ([]*authenticator.Info, error)
	New(userID string, spec authenticator.Spec, secret string) ([]*authenticator.Info, error)
	WithSecret(userID string, a *authenticator.Info, secret string) (changed bool, info *authenticator.Info, err error)
	CreateAll(userID string, ais []*authenticator.Info) error
	UpdateAll(userID string, ais []*authenticator.Info) error
	DeleteAll(userID string, ais []*authenticator.Info) error
	Authenticate(userID string, spec authenticator.Spec, state map[string]string, secret string) (*authenticator.Info, error)
	VerifySecret(userID string, a *authenticator.Info, state map[string]string, secret string) error
}

type OOBAuthenticatorProvider interface {
	GenerateCode(secret string, channel authn.AuthenticatorOOBChannel) string
	SendCode(
		channel authn.AuthenticatorOOBChannel,
		loginID *loginid.LoginID,
		code string,
		operation otp.OOBOperationType,
		stage otp.OOBAuthenticationStage,
	) error
}

type AnonymousIdentityProvider interface {
	ParseRequest(requestJWT string) (*anonymous.Identity, *anonymous.Request, error)
}

type ChallengeProvider interface {
	Consume(token string) (*challenge.Purpose, error)
}

type UserService interface {
	Get(id string) (*model.User, error)
	Create(userID string, metadata map[string]interface{}) (*user.User, error)
	AfterCreate(user *user.User, identities []*identity.Info, authenticators []*authenticator.Info) error
	UpdateLoginTime(user *model.User, lastLoginAt time.Time) error
}

type HookProvider interface {
	DispatchEvent(payload event.Payload, user *model.User) error
}

type SessionProvider interface {
	MakeSession(*authn.Attrs) (*session.IDPSession, string)
	Create(*session.IDPSession) error
}

type OAuthProviderFactory interface {
	NewOAuthProvider(alias string) sso.OAuthProvider
}

type Context struct {
	IsDryRun bool `wire:"-"`

	Database db.SQLExecutor
	Config   *config.AppConfig

	Identities           IdentityProvider
	Authenticators       AuthenticatorProvider
	AnonymousIdentities  AnonymousIdentityProvider
	OOBAuthenticators    OOBAuthenticatorProvider
	OAuthProviderFactory OAuthProviderFactory

	Challenges    ChallengeProvider
	Users         UserService
	Hooks         HookProvider
	Sessions      SessionProvider
	SessionCookie session.CookieDef
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
