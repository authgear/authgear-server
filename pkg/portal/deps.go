package portal

import (
	"github.com/google/wire"

	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	appresource "github.com/authgear/authgear-server/pkg/portal/appresource/factory"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/endpoint"
	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/portal/lib/plan"
	"github.com/authgear/authgear-server/pkg/portal/loader"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/task"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"

	// Import auth package to ensure correct content of registries
	_ "github.com/authgear/authgear-server/pkg/auth"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	deps.TaskDependencySet,

	service.DependencySet,
	adminauthz.DependencySet,
	clock.DependencySet,

	plan.DependencySet,
	globaldb.DependencySet,

	template.DependencySet,
	endpoint.DependencySet,

	wire.Bind(new(service.AuthzAdder), new(*adminauthz.Adder)),
	wire.Bind(new(service.CollaboratorServiceTaskQueue), new(*task.InProcessQueue)),
	wire.Bind(new(service.CollaboratorServiceEndpointsProvider), new(*endpoint.EndpointsProvider)),
	wire.Bind(new(service.CollaboratorServiceAdminAPIService), new(*service.AdminAPIService)),
	wire.Bind(new(service.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(service.AppPlanService), new(*plan.Service)),
	wire.Bind(new(service.AppResourceManagerFactory), new(*appresource.ManagerFactory)),

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

	transport.DependencySet,
	wire.Bind(new(transport.AdminAPIService), new(*service.AdminAPIService)),
	wire.Bind(new(transport.AdminAPIAuthzService), new(*service.AuthzService)),
	wire.Bind(new(transport.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(transport.SystemConfigProvider), new(*service.SystemConfigProvider)),

	appresource.DependencySet,
)
