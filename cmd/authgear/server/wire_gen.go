// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"context"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/queue"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// Injectors from wire.go:

func newConfigSourceController(p *deps.RootProvider, c context.Context) *configsource.Controller {
	config := p.ConfigSourceConfig
	factory := p.LoggerFactory
	localFSLogger := configsource.NewLocalFSLogger(factory)
	manager := p.BaseResources
	localFS := &configsource.LocalFS{
		Logger:        localFSLogger,
		BaseResources: manager,
		Config:        config,
	}
	databaseLogger := configsource.NewDatabaseLogger(factory)
	environmentConfig := p.EnvironmentConfig
	trustProxy := environmentConfig.TrustProxy
	clock := _wireSystemClockValue
	globalDatabaseCredentialsEnvironmentConfig := &environmentConfig.GlobalDatabase
	sqlBuilder := globaldb.NewSQLBuilder(globalDatabaseCredentialsEnvironmentConfig)
	storeFactory := configsource.NewStoreFactory(c, sqlBuilder)
	pool := p.DatabasePool
	databaseEnvironmentConfig := &environmentConfig.DatabaseConfig
	databaseHandleFactory := configsource.NewDatabaseHandleFactory(c, pool, globalDatabaseCredentialsEnvironmentConfig, databaseEnvironmentConfig, factory)
	resolveAppIDType := configsource.NewResolveAppIDTypeDomain()
	database := &configsource.Database{
		Logger:                databaseLogger,
		BaseResources:         manager,
		TrustProxy:            trustProxy,
		Config:                config,
		Clock:                 clock,
		StoreFactory:          storeFactory,
		DatabaseHandleFactory: databaseHandleFactory,
		DatabaseCredentials:   globalDatabaseCredentialsEnvironmentConfig,
		DatabaseConfig:        databaseEnvironmentConfig,
		ResolveAppIDType:      resolveAppIDType,
	}
	controller := configsource.NewController(config, localFS, database)
	return controller
}

var (
	_wireSystemClockValue = clock.NewSystemClock()
)

func newInProcessQueue(p *deps.AppProvider, e *executor.InProcessExecutor) *queue.InProcessQueue {
	handle := p.AppDatabase
	appContext := p.AppContext
	config := appContext.Config
	captureTaskContext := deps.ProvideCaptureTaskContext(config, appContext)
	inProcessQueue := &queue.InProcessQueue{
		Database:       handle,
		CaptureContext: captureTaskContext,
		Executor:       e,
	}
	return inProcessQueue
}
