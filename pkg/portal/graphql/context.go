package graphql

import (
	"context"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v72"

	"github.com/authgear/authgear-server/pkg/api/event"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/portal/appresource"
	"github.com/authgear/authgear-server/pkg/portal/appsecret"
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/smtp"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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
	LoadRawAppConfig(ctx context.Context, app *model.App) (*config.AppConfig, string, error)
	RenderSAMLEntityID(appID string) string

	Get(ctx context.Context, id string) (*model.App, error)
	GetAppList(ctx context.Context, userID string) ([]*model.AppListItem, error)
	Create(ctx context.Context, userID string, id string) (*model.App, error)
	UpdateResources(ctx context.Context, app *model.App, updates []appresource.Update) error
	UpdateResources0(ctx context.Context, app *model.App, updates []appresource.Update) error
	GetProjectQuota(ctx context.Context, userID string) (int, error)
	LoadAppSecretConfig(ctx context.Context, app *model.App, sessionInfo *apimodel.SessionInfo, token string) (*model.SecretConfig, string, error)
	LoadEffectiveSecretConfig(ctx context.Context, app *model.App) (*model.EffectiveSecretConfig, error)
	GenerateSecretVisitToken(ctx context.Context,
		app *model.App,
		sessionInfo *apimodel.SessionInfo,
		visitingSecrets []config.SecretKey,
	) (*appsecret.AppSecretVisitToken, error)
	GenerateTesterToken(ctx context.Context,
		app *model.App,
		returnURI string,
	) (*tester.TesterToken, error)
	LoadAppWebhookSecretMaterials(
		ctx context.Context,
		app *model.App) (*config.WebhookKeyMaterials, error)
}

type DomainService interface {
	ListDomains(ctx context.Context, appID string) ([]*apimodel.Domain, error)
	CreateCustomDomain(ctx context.Context, appID string, domain string) (*apimodel.Domain, error)
	DeleteDomain(ctx context.Context, appID string, id string) error
	VerifyDomain(ctx context.Context, appID string, id string) (*apimodel.Domain, error)
}

type CollaboratorService interface {
	GetCollaborator(ctx context.Context, id string) (*model.Collaborator, error)
	GetCollaboratorByAppAndUser(ctx context.Context, appID string, userID string) (*model.Collaborator, error)
	ListCollaborators(ctx context.Context, appID string) ([]*model.Collaborator, error)
	ListCollaboratorsByUser(ctx context.Context, userID string) ([]*model.Collaborator, error)
	DeleteCollaborator(ctx context.Context, c *model.Collaborator) error

	GetProjectOwnerCount(ctx context.Context, userID string) (int, error)

	GetInvitation(ctx context.Context, id string) (*model.CollaboratorInvitation, error)
	GetInvitationWithCode(ctx context.Context, id string) (*model.CollaboratorInvitation, error)
	ListInvitations(ctx context.Context, appID string) ([]*model.CollaboratorInvitation, error)
	DeleteInvitation(ctx context.Context, i *model.CollaboratorInvitation) error
	SendInvitation(ctx context.Context, appID string, inviteeEmail string) (*model.CollaboratorInvitation, error)
	AcceptInvitation(ctx context.Context, code string) (*model.Collaborator, error)
	CheckInviteeEmail(ctx context.Context, i *model.CollaboratorInvitation, actorID string) error
}

type AuthzService interface {
	CheckAccessOfViewer(ctx context.Context, appID string) (userID string, err error)
}

type SMTPService interface {
	SendTestEmail(ctx context.Context, app *model.App, options smtp.SendTestEmailOptions) (err error)
}

type SMSService interface {
	SendTestSMS(
		ctx context.Context,
		app *model.App,
		to string,
		webhookSecretLoader func(ctx context.Context) (*config.WebhookKeyMaterials, error),
		input model.SMSProviderConfigurationInput) error
}

type AppResourceManagerFactory interface {
	NewManagerWithAppContext(appContext *config.AppContext) *appresource.Manager
}

type TutorialService interface {
	Get(ctx context.Context, appID string) (*tutorial.Entry, error)
	RecordProgresses(ctx context.Context, appID string, ps []tutorial.Progress) (err error)
	Skip(ctx context.Context, appID string) (err error)
	SaveProjectWizardData(ctx context.Context, appID string, data interface{}) error
}

type OnboardService interface {
	SubmitOnboardEntry(ctx context.Context, entry model.OnboardEntry, actorID string) error
	CheckOnboardingSurveyCompletion(ctx context.Context, actorID string) (bool, error)
}

type AnalyticChartService interface {
	GetActiveUserChart(ctx context.Context, appID string, periodical string, rangeFrom time.Time, rangeTo time.Time) (*analytic.Chart, error)
	GetTotalUserCountChart(ctx context.Context, appID string, rangeFrom time.Time, rangeTo time.Time) (*analytic.Chart, error)
	GetSignupConversionRate(ctx context.Context, appID string, rangeFrom time.Time, rangeTo time.Time) (*analytic.SignupConversionRateData, error)
	GetSignupByMethodsChart(ctx context.Context, appID string, rangeFrom time.Time, rangeTo time.Time) (*analytic.Chart, error)
}

type StripeService interface {
	GenerateCustomerPortalSession(appID string, customerID string) (*stripe.BillingPortalSession, error)
	SetSubscriptionCancelAtPeriodEnd(stripeSubscriptionID string, cancelAtPeriodEnd bool) (*time.Time, error)

	FetchSubscriptionPlans(ctx context.Context) ([]*model.SubscriptionPlan, error)
	CreateCheckoutSession(ctx context.Context, appID string, customerEmail string, subscriptionPlan *model.SubscriptionPlan) (*libstripe.CheckoutSession, error)
	FetchCheckoutSession(ctx context.Context, checkoutSessionID string) (*libstripe.CheckoutSession, error)
	GetSubscriptionPlan(ctx context.Context, planName string) (*model.SubscriptionPlan, error)
	UpdateSubscription(ctx context.Context, stripeSubscriptionID string, subscriptionPlan *model.SubscriptionPlan) error
	PreviewUpdateSubscription(ctx context.Context, stripeSubscriptionID string, subscriptionPlan *model.SubscriptionPlan) (*model.SubscriptionUpdatePreview, error)
	GetLastPaymentError(ctx context.Context, stripeCustomerID string) (*stripe.Error, error)
	GetSubscription(ctx context.Context, stripeCustomerID string) (*stripe.Subscription, error)
	CancelSubscriptionImmediately(ctx context.Context, subscriptionID string) error
}

type SubscriptionService interface {
	GetSubscription(ctx context.Context, appID string) (*model.Subscription, error)
	CreateSubscriptionCheckout(ctx context.Context, stripeCheckoutSession *libstripe.CheckoutSession) (*model.SubscriptionCheckout, error)
	MarkCheckoutCompleted(ctx context.Context, appID string, stripCheckoutSessionID string, customerID string) error
	GetSubscriptionUsage(ctx context.Context,
		appID string,
		planName string,
		date time.Time,
		subscriptionPlans []*model.SubscriptionPlan,
	) (*model.SubscriptionUsage, error)
	UpdateAppPlan(ctx context.Context, appID string, planName string) error
	SetSubscriptionCancelledStatus(ctx context.Context, id string, cancelled bool, endedAt *time.Time) error
	GetLastProcessingCustomerID(ctx context.Context, appID string) (*string, error)
	MarkCheckoutExpired(ctx context.Context, appID string, customerID string) error
}

type UsageService interface {
	GetUsage(ctx context.Context,
		appID string,
		date time.Time,
	) (*model.Usage, error)
}

type DenoService interface {
	Check(ctx context.Context, snippet string) error
}

type AuditService interface {
	Log(ctx context.Context, app *model.App, payload event.NonBlockingPayload) error
}

type IPBlocklistService interface {
	CheckIP(ctx context.Context, ipAddress string, cidrs []string, countryCodes []string) bool
}

type Context struct {
	Request *http.Request

	GlobalDatabase *globaldb.Handle

	TrustProxy              config.TrustProxy
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
	SMSService           SMSService
	AppResMgrFactory     AppResourceManagerFactory
	AnalyticChartService AnalyticChartService
	TutorialService      TutorialService
	StripeService        StripeService
	SubscriptionService  SubscriptionService
	UsageService         UsageService
	DenoService          DenoService
	AuditService         AuditService
	OnboardService       OnboardService
	IPBlocklistService   IPBlocklistService
}

func WithContext(ctx context.Context, gqlContext *Context) context.Context {
	return graphqlutil.WithContext(ctx, gqlContext)
}

func GQLContext(ctx context.Context) *Context {
	return graphqlutil.GQLContext(ctx).(*Context)
}
