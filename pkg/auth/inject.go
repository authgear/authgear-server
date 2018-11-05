package auth

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
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

// Provide provides dependency instance by name
// nolint: gocyclo
func (m DependencyMap) Provide(dependencyName string, r *http.Request) interface{} {
	switch dependencyName {
	case "AuthContextGetter":
		return coreAuth.NewContextGetterWithContext(r.Context())
	case "TxContext":
		tConfig := config.GetTenantConfig(r)
		return db.NewTxContextWithContext(r.Context(), tConfig)
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
		return password.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), tConfig)),
			logging.CreateLogger(r, "provider_password", createLoggerMaskFormatter(r)),
			db.NewSafeTxContextWithContext(r.Context(), tConfig),
		)
	case "AnonymousAuthProvider":
		tConfig := config.GetTenantConfig(r)
		return anonymous.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), tConfig)),
			logging.CreateLogger(r, "provider_anonymous", createLoggerMaskFormatter(r)),
			db.NewSafeTxContextWithContext(r.Context(), tConfig),
		)
	case "HandlerLogger":
		return logging.CreateLogger(r, "handler", createLoggerMaskFormatter(r))
	case "UserProfileStore":
		tConfig := config.GetTenantConfig(r)
		switch tConfig.UserProfile.ImplName {
		default:
			panic("unrecgonized user profile store implementation: " + tConfig.UserProfile.ImplName)
		case "":
			// use auth default profile store
			return userprofile.NewSafeProvider(
				db.NewSQLBuilder("auth", tConfig.AppName),
				db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), tConfig)),
				logging.CreateLogger(r, "auth_user_profile", createLoggerMaskFormatter(r)),
				db.NewSafeTxContextWithContext(r.Context(), tConfig),
			)
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

func createLoggerMaskFormatter(r *http.Request) logrus.Formatter {
	tConfig := config.GetTenantConfig(r)
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}
