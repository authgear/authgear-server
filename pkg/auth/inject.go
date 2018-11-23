package auth

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/asset/fs"
	coreAudit "github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record/pq" // tolerant nextimportslint: record
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
	case "PasswordChecker":
		return &audit.PasswordChecker{
			// TODO:
			// from tConfig
			PwMinLength: 6,
		}
	case "PasswordAuthProvider":
		tConfig := config.GetTenantConfig(r)
		// TODO:
		// from tConfig
		authRecordKeys := [][]string{[]string{"email"}, []string{"username"}}
		return password.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), tConfig)),
			logging.CreateLogger(r, "provider_password", createLoggerMaskFormatter(r)),
			authRecordKeys,
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
		case "record":
			// use record based profile store
			roleStore := coreAuth.NewDefaultRoleStore(r.Context(), tConfig)
			recordStore := pq.NewSafeRecordStore(
				roleStore,
				// TODO: get from tconfig
				true,
				db.NewSQLBuilder("record", tConfig.AppName),
				db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), tConfig)),
				logging.CreateLogger(r, "record", createLoggerMaskFormatter(r)),
				db.NewSafeTxContextWithContext(r.Context(), tConfig),
			)
			// TODO: get from tConfig
			assetStore := fs.NewAssetStore("", "", "", true, logging.CreateLogger(r, "record", createLoggerMaskFormatter(r)))
			return userprofile.NewUserProfileRecordStore(
				tConfig.UserProfile.ImplStoreURL,
				tConfig.APIKey,
				logging.CreateLogger(r, "auth_user_profile", createLoggerMaskFormatter(r)),
				coreAuth.NewContextGetterWithContext(r.Context()),
				db.NewTxContextWithContext(r.Context(), tConfig),
				recordStore,
				assetStore,
			)
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
	case "WelcomeEmailSender":
		tConfig := config.GetTenantConfig(r)
		if !tConfig.WelcomeEmail.Enabled {
			return nil
		}

		return welcemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "TestWelcomeEmailSender":
		tConfig := config.GetTenantConfig(r)
		return welcemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.SMTP))
	default:
		return nil
	}
}

func createLoggerMaskFormatter(r *http.Request) logrus.Formatter {
	tConfig := config.GetTenantConfig(r)
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}
