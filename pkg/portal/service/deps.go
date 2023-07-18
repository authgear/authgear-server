package service

import (
	"github.com/authgear/authgear-server/pkg/portal/appsecret"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	appsecret.DependencySet,
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
	wire.Struct(new(NFTService), "*"),
	wire.Struct(new(AuditService), "*"),
	NewConfigServiceLogger,
	NewAppServiceLogger,
	NewKubernetesLogger,

	wire.Bind(new(AppAuthzService), new(*AuthzService)),
	wire.Bind(new(AppConfigService), new(*ConfigService)),
	wire.Bind(new(CollaboratorAppConfigService), new(*ConfigService)),
	wire.Bind(new(AuthzConfigService), new(*ConfigService)),
	wire.Bind(new(AuthzCollaboratorService), new(*CollaboratorService)),
	wire.Bind(new(DomainConfigService), new(*ConfigService)),
	wire.Bind(new(AppSecretVisitTokenStore), new(*appsecret.AppSecretVisitTokenStoreImpl)),
	wire.Bind(new(AppDefaultDomainService), new(*DefaultDomainService)),
	wire.Bind(new(AdminAPIDefaultDomainService), new(*DefaultDomainService)),
	wire.Bind(new(DefaultDomainDomainService), new(*DomainService)),
)
