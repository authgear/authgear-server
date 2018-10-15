package auth

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
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

func openDB(tConfig config.TenantConfiguration) func() (*sqlx.DB, error) {
	return func() (*sqlx.DB, error) {
		return sqlx.Open("postgres", tConfig.DBConnectionStr)
	}
}

// Provide provides dependency instance by name
// nolint: gocyclo
func (m DependencyMap) Provide(dependencyName string, r *http.Request) interface{} {
	switch dependencyName {
	case "AuthContextGetter":
		return coreAuth.NewContextGetterWithContext(r.Context())
	case "TxContext":
		tConfig := config.GetTenantConfig(r)
		return db.NewTxContextWithContext(r.Context(), openDB(tConfig))
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
		maskFormatter := logging.CreateMaskFormatter(sensitiveLoggerValues(r), &logrus.TextFormatter{})
		return password.NewProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), openDB(tConfig))),
			logging.CreateLogger(r, "provider_password", maskFormatter),
		)
	case "HandlerLogger":
		maskFormatter := logging.CreateMaskFormatter(sensitiveLoggerValues(r), &logrus.TextFormatter{})
		return logging.CreateLogger(r, "handler", maskFormatter)
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
	case "RoleStore":
		tConfig := config.GetTenantConfig(r)
		return coreAuth.NewDefaultRoleStore(r.Context(), tConfig)
	case "AuditTrail":
		tConfig := config.GetTenantConfig(r)
		trail, err := coreAudit.NewTrail(tConfig.UserAudit.Enabled, tConfig.UserAudit.TrailHandlerURL, r)
		if err != nil {
			panic(err)
		}
		return trail
	default:
		return nil
	}
}

func sensitiveLoggerValues(r *http.Request) []string {
	tConfig := config.GetTenantConfig(r)
	return []string{
		tConfig.APIKey,
		tConfig.MasterKey,
	}
}
