package graphql

import (
	"context"
	"time"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/smtp"
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
	UpdateResources(app *model.App, updates []appresource.Update) error
	GetMaxOwnedApps(userID string) (int, error)
	LoadRawAppConfig(app *model.App) (*config.AppConfig, error)
	LoadAppSecretConfig(app *model.App, sessionInfo *apimodel.SessionInfo) (*model.SecretConfig, error)
}

type DomainService interface {
	ListDomains(appID string) ([]*model.Domain, error)
	CreateCustomDomain(appID string, domain string) (*model.Domain, error)
	DeleteDomain(appID string, id string) error
	VerifyDomain(appID string, id string) (*model.Domain, error)
}

type CollaboratorService interface {
	GetCollaborator(id string) (*model.Collaborator, error)
	GetCollaboratorByAppAndUser(appID string, userID string) (*model.Collaborator, error)
	ListCollaborators(appID string) ([]*model.Collaborator, error)
	ListCollaboratorsByUser(userID string) ([]*model.Collaborator, error)
	DeleteCollaborator(c *model.Collaborator) error

	GetInvitation(id string) (*model.CollaboratorInvitation, error)
	GetInvitationWithCode(id string) (*model.CollaboratorInvitation, error)
	ListInvitations(appID string) ([]*model.CollaboratorInvitation, error)
	DeleteInvitation(i *model.CollaboratorInvitation) error
	SendInvitation(appID string, inviteeEmail string) (*model.CollaboratorInvitation, error)
	AcceptInvitation(code string) (*model.Collaborator, error)
	CheckInviteeEmail(i *model.CollaboratorInvitation, actorID string) error
}

type AuthzService interface {
	CheckAccessOfViewer(appID string) (userID string, err error)
}

type SMTPService interface {
	SendTestEmail(app *model.App, options smtp.SendTestEmailOptions) (err error)
}

type AppResourceManagerFactory interface {
	NewManagerWithAppContext(appContext *config.AppContext) *appresource.Manager
}

type AnalyticChartService interface {
	GetActiveUserChat(appID string, periodical string, rangeFrom time.Time, rangeTo time.Time) (*analytic.Chart, error)
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

	AuthzService         AuthzService
	AppService           AppService
	DomainService        DomainService
	CollaboratorService  CollaboratorService
	SMTPService          SMTPService
	AppResMgrFactory     AppResourceManagerFactory
	AnalyticChartService AnalyticChartService
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
