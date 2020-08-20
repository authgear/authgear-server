package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/admin/utils"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type UserLoader interface {
	Get(id string) *utils.Lazy
	QueryPage(args loader.PageArgs) (*loader.PageResult, error)
}

type IdentityLoader interface {
	Get(ref *identity.Ref) *utils.Lazy
	List(userID string) *utils.Lazy
}

type AuthenticatorLoader interface {
	Get(ref *authenticator.Ref) *utils.Lazy
	List(userID string) *utils.Lazy
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
