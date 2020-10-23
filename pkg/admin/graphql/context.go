package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserLoader interface {
	graphqlutil.DataLoaderInterface
}

type IdentityLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuthenticatorLoader interface {
	graphqlutil.DataLoaderInterface
}

type UserFacade interface {
	Get(userID string) (*user.User, error)
	QueryPage(args graphqlutil.PageArgs) (*graphqlutil.PageResult, error)
	Create(identityDef model.IdentityDef, password string) (string, error)
	ResetPassword(id string, password string) error
}

type IdentityFacade interface {
	Get(ref *identity.Ref) (*identity.Info, error)
	List(userID string) ([]*identity.Ref, error)
	Remove(identityInfo *identity.Info) error
	Create(userID string, identityDef model.IdentityDef, password string) (*identity.Ref, error)
}

type AuthenticatorFacade interface {
	Get(ref *authenticator.Ref) (*authenticator.Info, error)
	List(userID string) ([]*authenticator.Ref, error)
	Remove(authenticatorInfo *authenticator.Info) error
}

type VerificationFacade interface {
	Get(userID string) ([]model.Claim, error)
	SetVerified(userID string, claimName string, claimValue string, isVerified bool) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("admin-graphql")} }

type Context struct {
	GQLLogger Logger

	Users          UserLoader
	Identities     IdentityLoader
	Authenticators AuthenticatorLoader

	UserFacade          UserFacade
	IdentityFacade      IdentityFacade
	AuthenticatorFacade AuthenticatorFacade
	VerificationFacade  VerificationFacade
}

func (c *Context) Logger() *log.Logger {
	return c.GQLLogger.Logger
}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return graphqlutil.WithContext(ctx, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return graphqlutil.GQLContext(ctx).(*Context)
}
