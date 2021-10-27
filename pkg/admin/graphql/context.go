package graphql

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/admin/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
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

type AuditLogLoader interface {
	graphqlutil.DataLoaderInterface
}

type AuditLogFacade interface {
	QueryPage(opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
}

type UserFacade interface {
	ListPage(sortOption libuser.SortOption, args graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	SearchPage(searchKeyword string, sortOption libuser.SortOption, args graphqlutil.PageArgs) ([]apimodel.PageItemRef, *graphqlutil.PageResult, error)
	Create(identityDef model.IdentityDef, password string) (string, error)
	ResetPassword(id string, password string) error
	SetDisabled(id string, isDisabled bool, reason *string) error
	Delete(id string) error
	UpdateStandardAttributes(id string, stdAttrs map[string]interface{}) error
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
	DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
}

type SessionFacade interface {
	List(userID string) ([]session.Session, error)
	Get(id string) (session.Session, error)
	Revoke(id string) error
	RevokeAll(userID string) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("admin-graphql")} }

type Context struct {
	GQLLogger Logger

	Users          UserLoader
	Identities     IdentityLoader
	Authenticators AuthenticatorLoader
	AuditLogs      AuditLogLoader

	UserFacade          UserFacade
	AuditLogFacade      AuditLogFacade
	IdentityFacade      IdentityFacade
	AuthenticatorFacade AuthenticatorFacade
	VerificationFacade  VerificationFacade
	SessionFacade       SessionFacade
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
