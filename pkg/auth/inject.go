package auth

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/server/audit"
)

type DependencyMap struct{}

func NewDependencyMap() DependencyMap {
	return DependencyMap{}
}

func (m DependencyMap) Provide(dependencyName string, r *http.Request) interface{} {
	switch dependencyName {
	case "TokenStore":
		tConfig := config.GetTenantConfig(r)
		return coreAuth.NewDefaultTokenStore(r.Context(), tConfig)
	case "AuthInfoStore":
		tConfig := config.GetTenantConfig(r)
		return coreAuth.NewDefaultAuthInfoStore(r.Context(), tConfig)
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
	case "PasswordAuthProvider":
		tConfig := config.GetTenantConfig(r)
		return password.NewProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(r.Context(), "postgres", tConfig.DBConnectionStr),
			logging.CreateLogger(r, "provider_password"),
		)
	case "HandlerLogger":
		return logging.CreateLogger(r, "handler")
	case "UserProfileStore":
		tConfig := config.GetTenantConfig(r)
		switch tConfig.UserProfile.ImplName {
		default:
			panic("unrecgonized user profile store implementation: " + tConfig.UserProfile.ImplName)
		case "":
			return nil
			// case "skygear":
			// 	return XXX
		}
	default:
		return nil
	}
}
