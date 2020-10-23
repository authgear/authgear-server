package graphql

import (
	"context"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/util/resources"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserLoader interface {
	graphqlutil.DataLoaderInterface
}

type AppLoader interface {
	graphqlutil.DataLoaderInterface
}

type DomainLoader interface {
	graphqlutil.DataLoaderInterface
}

type CollaboratorLoader interface {
	graphqlutil.DataLoaderInterface
}

type CollaboratorInvitationLoader interface {
	graphqlutil.DataLoaderInterface
}

type AppService interface {
	Get(id string) (*model.App, error)
	List(userID string) ([]*model.App, error)
	Create(userID string, id string) error
	UpdateResources(app *model.App, updates []resources.Update) error
}

type DomainService interface {
	ListDomains(appID string) ([]*model.Domain, error)
	CreateCustomDomain(appID string, domain string) (*model.Domain, error)
	DeleteDomain(appID string, id string) error
	VerifyDomain(appID string, id string) (*model.Domain, error)
}

type CollaboratorService interface {
	GetCollaborator(id string) (*model.Collaborator, error)
	ListCollaborators(appID string) ([]*model.Collaborator, error)
	DeleteCollaborator(c *model.Collaborator) error

	GetInvitation(id string) (*model.CollaboratorInvitation, error)
	ListInvitations(appID string) ([]*model.CollaboratorInvitation, error)
	DeleteInvitation(i *model.CollaboratorInvitation) error
	SendInvitation(appID string, inviteeEmail string) (*model.CollaboratorInvitation, error)
	AcceptInvitation(code string) (*model.Collaborator, error)
}

type AuthzService interface {
	CheckAccessOfViewer(appID string) (userID string, err error)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("portal-graphql")} }

type Context struct {
	GQLLogger Logger

	Users                   UserLoader
	Apps                    AppLoader
	Domains                 DomainLoader
	Collaborators           CollaboratorLoader
	CollaboratorInvitations CollaboratorInvitationLoader

	AuthzService        AuthzService
	AppService          AppService
	DomainService       DomainService
	CollaboratorService CollaboratorService
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
