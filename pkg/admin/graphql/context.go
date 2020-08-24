package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type UserLoader interface {
	Get(id string) *graphqlutil.Lazy
	QueryPage(args graphqlutil.PageArgs) (*graphqlutil.PageResult, error)
}

type IdentityLoader interface {
	Get(ref *identity.Ref) *graphqlutil.Lazy
	List(userID string) *graphqlutil.Lazy
}

type AuthenticatorLoader interface {
	Get(ref *authenticator.Ref) *graphqlutil.Lazy
	List(userID string) *graphqlutil.Lazy
}

type Context struct {
	Users          UserLoader
	Identities     IdentityLoader
	Authenticators AuthenticatorLoader
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return context.WithValue(ctx, contextKey, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return ctx.Value(contextKey).(*Context)
}
