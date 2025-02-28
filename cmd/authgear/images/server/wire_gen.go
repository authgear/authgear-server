// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/google/wire"
)

// Injectors from wire.go:

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
	environmentConfig := &p.EnvironmentConfig
	config := environmentConfig.ConfigSource
	factory := p.LoggerFactory
	localFSLogger := configsource.NewLocalFSLogger(factory)
	manager := p.BaseResources
	localFS := &configsource.LocalFS{
		Logger:        localFSLogger,
		BaseResources: manager,
		Config:        config,
	}
	databaseLogger := configsource.NewDatabaseLogger(factory)
	trustProxy := environmentConfig.TrustProxy
	clock := _wireSystemClockValue
	globalDatabaseCredentialsEnvironmentConfig := environmentConfig.GlobalDatabase
	sqlBuilder := globaldb.NewSQLBuilder(globalDatabaseCredentialsEnvironmentConfig)
	configSourceStoreFactory := configsource.NewConfigSourceStoreStoreFactory(sqlBuilder)
	planStoreFactory := configsource.NewPlanStoreStoreFactory(sqlBuilder)
	pool := p.DatabasePool
	databaseEnvironmentConfig := environmentConfig.DatabaseConfig
	databaseHandleFactory := configsource.NewDatabaseHandleFactory(pool, globalDatabaseCredentialsEnvironmentConfig, databaseEnvironmentConfig, factory)
	resolveAppIDType := configsource.NewResolveAppIDTypePath()
	database := &configsource.Database{
		Logger:                   databaseLogger,
		BaseResources:            manager,
		TrustProxy:               trustProxy,
		Config:                   config,
		Clock:                    clock,
		ConfigSourceStoreFactory: configSourceStoreFactory,
		PlanStoreFactory:         planStoreFactory,
		DatabaseHandleFactory:    databaseHandleFactory,
		DatabaseCredentials:      globalDatabaseCredentialsEnvironmentConfig,
		DatabaseConfig:           databaseEnvironmentConfig,
		ResolveAppIDType:         resolveAppIDType,
	}
	controller := configsource.NewController(config, localFS, database)
	return controller
}

var (
	_wireSystemClockValue = clock.NewSystemClock()
)

// wire.go:

var configSourceConfigDependencySet = wire.NewSet(globaldb.DependencySet, clock.DependencySet, wire.FieldsOf(new(*deps.RootProvider),
	"EnvironmentConfig",
	"LoggerFactory",
	"DatabasePool",
	"BaseResources",
), wire.FieldsOf(new(*config.EnvironmentConfig),
	"TrustProxy",
	"ConfigSource",
	"GlobalDatabase",
	"DatabaseConfig",
),
)
