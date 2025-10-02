package elasticsearch

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/wire"
	"github.com/kelseyhightower/envconfig"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/globalredis"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

type CmdAppID string
type CmdDBCredential config.DatabaseCredentials

func NewEmptyConfig(
	pool *db.Pool,
	databaseCredentials *CmdDBCredential,
	appID CmdAppID,
) *config.Config {
	dbCred := config.DatabaseCredentials(*databaseCredentials)
	featureConfig := &config.FeatureConfig{}
	config.PopulateFeatureConfigDefaultValues(featureConfig)

	appConfig := &config.AppConfig{
		ID: config.AppID(appID),
	}
	config.PopulateDefaultValues(appConfig)

	return &config.Config{
		AppConfig: appConfig,
		SecretConfig: &config.SecretConfig{
			Secrets: []config.SecretItem{
				{
					Key:  config.DatabaseCredentialsKey,
					Data: &dbCred,
				},
			},
		},
		FeatureConfig: featureConfig,
	}
}

func NewEnvConfig(dbCredentials *CmdDBCredential) *config.EnvironmentConfig {
	cfg := &config.EnvironmentConfig{}

	err := envconfig.Process("", cfg)
	if err != nil {
		panic(fmt.Errorf("cannot load server config: %w", err))
	}

	cfg.GlobalDatabase = config.GlobalDatabaseCredentialsEnvironmentConfig{
		DatabaseURL:    dbCredentials.DatabaseURL,
		DatabaseSchema: dbCredentials.DatabaseSchema,
	}
	cfg.DatabaseConfig = *config.NewDefaultDatabaseEnvironmentConfig()

	return cfg
}

type NilResourceManager interface {
	Read(ctx context.Context, desc resource.Descriptor, view resource.View) (interface{}, error)
	AssetName(key string) (name string, err error)
}

func NewNilResourceManager() NilResourceManager {
	return nil
}

func NewNilRedis() *appredis.Handle {
	return nil
}

func NewNilGlobalRedis() *globalredis.Handle {
	return nil
}

func NewNilRequest() *http.Request {
	return nil
}

func ProvideRemoteIP() httputil.RemoteIP {
	return "127.0.0.1"
}

func ProvideHTTPHost() httputil.HTTPHost {
	return ""
}

func ProvideHTTPProto() httputil.HTTPProto {
	return "http"
}

var DependencySet = wire.NewSet(
	ProvideRemoteIP,
	ProvideHTTPHost,
	ProvideHTTPProto,
	NewNilGlobalRedis,
	NewNilRedis,
	NewEnvConfig,
	NewNilRequest,
	NewEmptyConfig,
	NewReindexedTimestamps,
	globaldb.DependencySet,
	appdb.NewHandle,
	clock.DependencySet,
	deps.EnvConfigDeps,
	deps.CommonDependencySet,
	wire.Struct(new(configsource.Store), "*"),
	wire.Struct(new(AppLister), "*"),
	wire.Struct(new(Reindexer), "*"),

	NewNilResourceManager,
	wire.Bind(new(loginid.ResourceManager), new(NilResourceManager)),
	wire.Bind(new(template.ResourceManager), new(NilResourceManager)),
	wire.Bind(new(web.ResourceManager), new(NilResourceManager)),
	wire.Bind(new(web.EmbeddedResourceManager), new(NilResourceManager)),
)
