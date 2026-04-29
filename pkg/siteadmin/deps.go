package siteadmin

import (
	"github.com/google/wire"

	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/config/plan"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	usagepkg "github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	siteadminservice "github.com/authgear/authgear-server/pkg/siteadmin/service"
	"github.com/authgear/authgear-server/pkg/siteadmin/transport"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	clock.DependencySet,
	// Use a dedicated connection pool for siteadmin (ConnectionPurposeSiteadminGlobal)
	// so siteadmin traffic cannot exhaust the global DB connections of other components
	// (portal, APIs, etc.).
	newSiteadminGlobalHandle,
	globaldb.NewSQLExecutor,
	globaldb.NewSQLBuilder,
	globalredis.DependencySet,
	session.DependencySet,
	transport.DependencySet,
	wire.FieldsOf(new(*config.EnvironmentConfig), "CORSAllowedOrigins"),
	wire.Struct(new(CORSMatcher), "*"),
	wire.Bind(new(middleware.CORSOriginMatcher), new(*CORSMatcher)),
	wire.Struct(new(portalservice.CollaboratorService), "SQLBuilder", "SQLExecutor", "GlobalDatabase"),
	wire.Bind(new(transport.AuthzCollaboratorService), new(*portalservice.CollaboratorService)),

	// siteadmin service layer
	siteadminservice.DependencySet,
	wire.Bind(new(siteadminservice.AppServiceDatabase), new(*globaldb.Handle)),

	// adminauthz.Adder satisfies portalservice.AuthzAdder
	wire.Struct(new(adminauthz.Adder), "Clock"),
	wire.Bind(new(portalservice.AuthzAdder), new(*adminauthz.Adder)),

	// DefaultDomainService (partial — only fields needed for GetLatestAppHost)
	wire.Struct(new(portalservice.DefaultDomainService), "AppHostSuffixes", "AppConfig"),
	wire.Bind(new(portalservice.AdminAPIDefaultDomainService), new(*portalservice.DefaultDomainService)),

	// AdminAPIService wires up SelfDirector used by AppService
	wire.Struct(new(portalservice.AdminAPIService), "AuthgearConfig", "AdminAPIConfig", "ConfigSource", "AuthzAdder", "DefaultDomains"),
	wire.Bind(new(siteadminservice.SiteAdminAdminAPI), new(*portalservice.AdminAPIService)),

	// configsource.Store satisfies AppServiceConfigSourceStore
	wire.Struct(new(configsource.Store), "*"),
	wire.Bind(new(siteadminservice.AppServiceConfigSourceStore), new(*configsource.Store)),

	// Audit DB (optional — nil when not configured)
	auditdb.DependencySet,
	auditdb.NewReadHandle,
	wire.Struct(new(analytic.AuditDBReadStore), "*"),

	// usage.GlobalDBStore satisfies UsageServiceGlobalDBStore
	wire.Struct(new(usagepkg.GlobalDBStore), "SQLBuilder", "SQLExecutor"),
	wire.Bind(new(siteadminservice.UsageServiceGlobalDBStore), new(*usagepkg.GlobalDBStore)),

	// plan.Store satisfies PlanServicePlanStore
	wire.Struct(new(plan.Store), "SQLBuilder", "SQLExecutor", "Clock"),
	wire.Bind(new(siteadminservice.PlanServicePlanStore), new(*plan.Store)),

	// PlanService additional bindings
	wire.Bind(new(siteadminservice.PlanServiceGlobalDatabase), new(*globaldb.Handle)),
	wire.Bind(new(siteadminservice.PlanServiceConfigSourceStore), new(*configsource.Store)),
	wire.Bind(new(siteadminservice.PlanServiceOwnerStore), new(*siteadminservice.AppOwnerStore)),

	// transport bindings
	wire.Bind(new(transport.AppsListService), new(*siteadminservice.AppService)),
	wire.Bind(new(transport.AppGetService), new(*siteadminservice.AppService)),
	wire.Bind(new(transport.CollaboratorsListService), new(*siteadminservice.CollaboratorService)),
	wire.Bind(new(transport.CollaboratorAddService), new(*siteadminservice.CollaboratorService)),
	wire.Bind(new(transport.CollaboratorRemoveService), new(*siteadminservice.CollaboratorService)),
	wire.Bind(new(transport.MessagingUsageService), new(*siteadminservice.UsageService)),
	wire.Bind(new(transport.MonthlyActiveUsersUsageService), new(*siteadminservice.UsageService)),
	wire.Bind(new(transport.PlansListService), new(*siteadminservice.PlanService)),
	wire.Bind(new(transport.AppPlanChangeService), new(*siteadminservice.PlanService)),
	wire.Bind(new(transport.CollaboratorPromoteService), new(*siteadminservice.CollaboratorService)),
)
