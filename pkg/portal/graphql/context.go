package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/resources"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type AppLoader interface {
	Get(id string) *graphqlutil.Lazy
	List(userID string) *graphqlutil.Lazy

	Create(userID string, id string) *graphqlutil.Lazy
	UpdateResources(app *model.App, updates []resources.Update) *graphqlutil.Lazy
}

type DomainLoader interface {
	ListDomains(appID string) *graphqlutil.Lazy
	CreateDomain(appID string, domain string) *graphqlutil.Lazy
	DeleteDomain(appID string, id string) *graphqlutil.Lazy
	VerifyDomain(appID string, id string) *graphqlutil.Lazy
}

type CollaboratorLoader interface {
	ListCollaborators(appID string) *graphqlutil.Lazy
	DeleteCollaborator(id string) *graphqlutil.Lazy

	ListInvitations(appID string) *graphqlutil.Lazy
	DeleteInvitation(id string) *graphqlutil.Lazy
	SendInvitation(appID string, inviteeEmail string) *graphqlutil.Lazy
	AcceptInvitation(code string) *graphqlutil.Lazy
}

type UserLoader interface {
	graphqlutil.DataLoaderInterface
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("portal-graphql")} }

type Context struct {
	Users         UserLoader
	GQLLogger     Logger
	Apps          AppLoader
	Domains       DomainLoader
	Collaborators CollaboratorLoader
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
