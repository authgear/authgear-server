package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/server/audit"
	"github.com/skygeario/skygear-server/pkg/server/logging"
)

type DependencyMap struct{}

func NewDependencyMap() DependencyMap {
	return DependencyMap{}
}

func (m DependencyMap) Provide(dependencyName string, ctx context.Context, tConfig config.TenantConfiguration) interface{} {
	switch dependencyName {
	case "TokenStore":
		return coreAuth.NewDefaultTokenStore(ctx, tConfig)
	case "AuthInfoStore":
		return coreAuth.NewDefaultAuthInfoStore(ctx, tConfig)
	case "AuthDataChecker":
		return &dependency.DefaultAuthDataChecker{
			//TODO:
			// from tConfig
			AuthRecordKeys: [][]string{[]string{"email"}, []string{"username"}},
		}
	case "PasswordChecker":
		return &audit.PasswordChecker{
			// TODO:
			// from tConfig
			PwMinLength: 6,
		}
	case "AuthPrincipalStore":
		return principal.NewStore(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, "postgres", tConfig.DBConnectionStr),
			logging.CreateLogger(ctx, "auth_principal"),
		)
	default:
		return nil
	}
}
