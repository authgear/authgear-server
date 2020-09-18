package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserLoader interface {
	Get(id string) *graphqlutil.Lazy
	QueryPage(args graphqlutil.PageArgs) (*graphqlutil.PageResult, error)
}

type IdentityLoader interface {
	Get(ref *identity.Ref) *graphqlutil.Lazy
	List(userID string) *graphqlutil.Lazy

	Remove(identityInfo *identity.Info) *graphqlutil.Lazy
}

type AuthenticatorLoader interface {
	Get(ref *authenticator.Ref) *graphqlutil.Lazy
	List(userID string) *graphqlutil.Lazy
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("admin-graphql")} }

type Context struct {
	GQLLogger      Logger
	Users          UserLoader
	Identities     IdentityLoader
	Authenticators AuthenticatorLoader
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
