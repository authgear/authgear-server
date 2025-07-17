package cmdpricing

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

func NewGlobalDatabaseCredentials(dbCredentials *config.DatabaseCredentials) *config.GlobalDatabaseCredentialsEnvironmentConfig {
	return &config.GlobalDatabaseCredentialsEnvironmentConfig{
		DatabaseURL:    dbCredentials.DatabaseURL,
		DatabaseSchema: dbCredentials.DatabaseSchema,
	}
}

var DependencySet = wire.NewSet(
	config.NewDefaultDatabaseEnvironmentConfig,
	NewGlobalDatabaseCredentials,
	globaldb.DependencySet,
	clock.DependencySet,
	NewClientAPI,
	wire.Struct(new(configsource.Store), "*"),
	wire.Struct(new(StripeService), "*"),
)
