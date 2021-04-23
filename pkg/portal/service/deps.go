package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AppService), "*"),
	wire.Struct(new(AdminAPIService), "*"),
	wire.Struct(new(AuthzService), "*"),
	wire.Struct(new(ConfigService), "*"),
	wire.Struct(new(Kubernetes), "*"),
	wire.Struct(new(DomainService), "*"),
	wire.Struct(new(CollaboratorService), "*"),
	wire.Struct(new(SystemConfigProvider), "*"),
	NewConfigServiceLogger,
	NewAppServiceLogger,
	NewKubernetesLogger,

	wire.Bind(new(AppAuthzService), new(*AuthzService)),
	wire.Bind(new(AppConfigService), new(*ConfigService)),
	wire.Bind(new(AppAdminAPIService), new(*AdminAPIService)),
	wire.Bind(new(AppDomainService), new(*DomainService)),
	wire.Bind(new(AuthzConfigService), new(*ConfigService)),
	wire.Bind(new(AuthzCollaboratorService), new(*CollaboratorService)),
	wire.Bind(new(DomainConfigService), new(*ConfigService)),
)
