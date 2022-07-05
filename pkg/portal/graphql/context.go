package graphql

import (
	"context"
	"time"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/smtp"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/stripe/stripe-go/v72"
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

type TutorialService interface {
	Get(appID string) (*tutorial.Entry, error)
	RecordProgresses(appID string, ps []tutorial.Progress) (err error)
	Skip(appID string) (err error)
}

type AnalyticChartService interface {
	GetActiveUserChart(appID string, periodical string, rangeFrom time.Time, rangeTo time.Time) (*analytic.Chart, error)
	GetTotalUserCountChart(appID string, rangeFrom time.Time, rangeTo time.Time) (*analytic.Chart, error)
	GetSignupConversionRate(appID string, rangeFrom time.Time, rangeTo time.Time) (*analytic.SignupConversionRateData, error)
	GetSignupByMethodsChart(appID string, rangeFrom time.Time, rangeTo time.Time) (*analytic.Chart, error)
}

type StripeService interface {
	FetchSubscriptionPlans() ([]*model.SubscriptionPlan, error)
	CreateCheckoutSession(appID string, customerEmail string, subscriptionPlan *model.SubscriptionPlan) (*libstripe.CheckoutSession, error)
	FetchCheckoutSession(checkoutSessionID string) (*libstripe.CheckoutSession, error)
	GetSubscriptionPlan(planName string) (*model.SubscriptionPlan, error)
	GenerateCustomerPortalSession(appID string, customerID string) (*stripe.BillingPortalSession, error)
	UpdateSubscription(stripeSubscriptionID string, subscriptionPlan *model.SubscriptionPlan) error
	PreviewUpdateSubscription(stripeSubscriptionID string, subscriptionPlan *model.SubscriptionPlan) (*model.SubscriptionUpdatePreview, error)
}

type SubscriptionService interface {
	GetSubscription(appID string) (*model.Subscription, error)
	CreateSubscriptionCheckout(stripeCheckoutSession *libstripe.CheckoutSession) (*model.SubscriptionCheckout, error)
	UpdateSubscriptionCheckoutStatusAndCustomerID(appID string, stripCheckoutSessionID string, status model.SubscriptionCheckoutStatus, customerID string) error
	GetSubscriptionUsage(
		appID string,
		planName string,
		date time.Time,
		subscriptionPlans []*model.SubscriptionPlan,
	) (*model.SubscriptionUsage, error)
	GetIsProcessingSubscription(appID string) (bool, error)
	UpdateAppPlan(appID string, planName string) error
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
	TutorialService      TutorialService
	StripeService        StripeService
	SubscriptionService  SubscriptionService
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
