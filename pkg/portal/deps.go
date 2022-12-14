package portal

import (
	"time"

	"github.com/google/wire"

	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/tutorial"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	appresource "github.com/authgear/authgear-server/pkg/portal/appresource/factory"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/endpoint"
	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/portal/lib/plan"
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/portal/loader"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/smtp"
	"github.com/authgear/authgear-server/pkg/portal/task"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"

	// Import auth package to ensure correct content of registries
	_ "github.com/authgear/authgear-server/pkg/auth"
)

func ProvideEmptyAppID() config.AppID {
	return config.AppID("")
}

func ProvideDenoClient(endpoint config.DenoEndpoint, logger hook.Logger) *hook.DenoClient {
	return &hook.DenoClient{
		Endpoint:   string(endpoint),
		HTTPClient: httputil.NewExternalClient(5 * time.Second),
		Logger:     logger,
	}
}

var denoDependencySet = wire.NewSet(
	ProvideDenoClient,
	hook.NewLogger,
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	deps.TaskDependencySet,

	service.DependencySet,
	adminauthz.DependencySet,
	clock.DependencySet,

	globaldb.DependencySet,

	template.DependencySet,
	endpoint.DependencySet,

	smtp.DependencySet,

	auditdb.NewReadHandle,
	auditdb.DependencySet,
	analytic.DependencySet,
	configsource.DependencySet,

	usage.DependencySet,

	wire.Bind(new(service.AuthzAdder), new(*adminauthz.Adder)),
	wire.Bind(new(service.CollaboratorServiceTaskQueue), new(*task.InProcessQueue)),
	wire.Bind(new(service.CollaboratorServiceEndpointsProvider), new(*endpoint.EndpointsProvider)),
	wire.Bind(new(service.CollaboratorServiceAdminAPIService), new(*service.AdminAPIService)),
	wire.Bind(new(service.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(service.AppPlanService), new(*plan.Service)),
	wire.Bind(new(service.AppResourceManagerFactory), new(*appresource.ManagerFactory)),
	wire.Bind(new(service.SubscriptionConfigSourceStore), new(*configsource.Store)),
	wire.Bind(new(service.SubscriptionPlanStore), new(*plan.Store)),
	wire.Bind(new(service.UsageStore), new(*usage.GlobalDBStore)),

	loader.DependencySet,
	wire.Bind(new(loader.UserLoaderAdminAPIService), new(*service.AdminAPIService)),
	wire.Bind(new(loader.AppLoaderAppService), new(*service.AppService)),
	wire.Bind(new(loader.DomainLoaderDomainService), new(*service.DomainService)),
	wire.Bind(new(loader.CollaboratorLoaderCollaboratorService), new(*service.CollaboratorService)),
	wire.Bind(new(loader.AuthzService), new(*service.AuthzService)),

	graphql.DependencySet,
	wire.Bind(new(graphql.UserLoader), new(*loader.UserLoader)),
	wire.Bind(new(graphql.AppLoader), new(*loader.AppLoader)),
	wire.Bind(new(graphql.DomainLoader), new(*loader.DomainLoader)),
	wire.Bind(new(graphql.CollaboratorLoader), new(*loader.CollaboratorLoader)),
	wire.Bind(new(graphql.CollaboratorInvitationLoader), new(*loader.CollaboratorInvitationLoader)),
	wire.Bind(new(graphql.AuthzService), new(*service.AuthzService)),
	wire.Bind(new(graphql.AppService), new(*service.AppService)),
	wire.Bind(new(graphql.DomainService), new(*service.DomainService)),
	wire.Bind(new(graphql.CollaboratorService), new(*service.CollaboratorService)),
	wire.Bind(new(graphql.SMTPService), new(*smtp.Service)),
	wire.Bind(new(graphql.AppResourceManagerFactory), new(*appresource.ManagerFactory)),
	wire.Bind(new(graphql.AnalyticChartService), new(*analytic.ChartService)),
	wire.Bind(new(graphql.TutorialService), new(*tutorial.Service)),
	wire.Bind(new(graphql.StripeService), new(*libstripe.Service)),
	wire.Bind(new(graphql.SubscriptionService), new(*service.SubscriptionService)),
	wire.Bind(new(graphql.NFTService), new(*service.NFTService)),
	wire.Bind(new(graphql.DenoService), new(*hook.DenoClient)),

	transport.DependencySet,
	wire.Bind(new(transport.AdminAPIService), new(*service.AdminAPIService)),
	wire.Bind(new(transport.AdminAPIAuthzService), new(*service.AuthzService)),
	wire.Bind(new(transport.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(transport.SystemConfigProvider), new(*service.SystemConfigProvider)),
	wire.Bind(new(transport.StripeService), new(*libstripe.Service)),
	wire.Bind(new(transport.SubscriptionService), new(*service.SubscriptionService)),

	plan.DependencySet,
	wire.Bind(new(libstripe.PlanService), new(*plan.Service)),
	wire.Bind(new(libstripe.EndpointsProvider), new(*endpoint.EndpointsProvider)),

	appresource.DependencySet,

	tutorial.DependencySet,

	libstripe.DependencySet,

	denoDependencySet,
)
