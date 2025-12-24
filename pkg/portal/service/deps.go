package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/lib/tester"
	"github.com/authgear/authgear-server/pkg/portal/appsecret"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type HTTPClient struct {
	*http.Client
}

func NewHTTPClient() HTTPClient {
	return HTTPClient{
		httputil.NewExternalClient(5 * time.Second),
	}
}

var DependencySet = wire.NewSet(
	appsecret.DependencySet,
	tester.DependencySet,
	NewHTTPClient,
	wire.Struct(new(AppService), "*"),
	wire.Struct(new(AdminAPIService), "*"),
	wire.Struct(new(AuthzService), "*"),
	wire.Struct(new(ConfigService), "*"),
	wire.Struct(new(Kubernetes), "*"),
	wire.Struct(new(DomainService), "*"),
	wire.Struct(new(DefaultDomainService), "*"),
	wire.Struct(new(CollaboratorService), "*"),
	wire.Struct(new(SystemConfigProvider), "*"),
	wire.Struct(new(SubscriptionService), "*"),
	wire.Struct(new(UsageService), "*"),
	wire.Struct(new(AuditService), "*"),
	wire.Struct(new(OnboardService), "*"),
	wire.Struct(new(TokenService), "*"),

	wire.Bind(new(AppAuthzService), new(*AuthzService)),
	wire.Bind(new(AppConfigService), new(*ConfigService)),
	wire.Bind(new(CollaboratorAppConfigService), new(*ConfigService)),
	wire.Bind(new(AuthzConfigService), new(*ConfigService)),
	wire.Bind(new(AuthzCollaboratorService), new(*CollaboratorService)),
	wire.Bind(new(DomainConfigService), new(*ConfigService)),
	wire.Bind(new(AppSecretVisitTokenStore), new(*appsecret.AppSecretVisitTokenStoreImpl)),
	wire.Bind(new(AppTesterTokenStore), new(*tester.TesterStore)),
	wire.Bind(new(AppDefaultDomainService), new(*DefaultDomainService)),
	wire.Bind(new(AdminAPIDefaultDomainService), new(*DefaultDomainService)),
	wire.Bind(new(DefaultDomainDomainService), new(*DomainService)),
	wire.Bind(new(AuditServiceAppService), new(*AppService)),
)

type NoopAttributesService struct{}

func (*NoopAttributesService) UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error {
	return nil
}

func (*NoopAttributesService) UpdateAllCustomAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error {
	return nil
}

type NoopRolesAndGroupsService struct{}

func (*NoopRolesAndGroupsService) ResetUserRole(ctx context.Context, options *rolesgroups.ResetUserRoleOptions) error {
	return nil
}

func (*NoopRolesAndGroupsService) ResetUserGroup(ctx context.Context, options *rolesgroups.ResetUserGroupOptions) error {
	return nil
}

var AuthgearDependencySet = wire.NewSet(
	wire.FieldsOf(new(*model.App),
		"Context",
	),
	wire.FieldsOf(new(*config.AppContext),
		"Resources",
		"Config",
	),
	wire.Value(&NoopAttributesService{}),
	wire.Value(&NoopRolesAndGroupsService{}),
	auditdb.NewWriteHandle,

	deps.ConfigDeps,
	clock.DependencySet,
	auditdb.DependencySet,
	audit.DependencySet,

	hook.DependencySet,
	wire.Bind(new(hook.ResourceManager), new(*resource.Manager)),
	wire.Bind(new(hook.StandardAttributesServiceNoEvent), new(*NoopAttributesService)),
	wire.Bind(new(hook.CustomAttributesServiceNoEvent), new(*NoopAttributesService)),
	wire.Bind(new(hook.RolesAndGroupsServiceNoEvent), new(*NoopRolesAndGroupsService)),
)
