package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify/verifycode"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/customtoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/asset/fs"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record/pq" // tolerant nextimportslint: record
)

type DependencyMap struct {
	AsyncTaskExecutor *async.Executor
}

// Provide provides dependency instance by name
// nolint: gocyclo, golint
func (m DependencyMap) Provide(
	dependencyName string,
	ctx context.Context,
	requestID string,
	tConfig config.TenantConfiguration,
) interface{} {
	switch dependencyName {
	case "AuthContextGetter":
		return coreAuth.NewContextGetterWithContext(ctx)
	case "TxContext":
		return db.NewTxContextWithContext(ctx, tConfig)
	case "TokenStore":
		return coreAuth.NewDefaultTokenStore(ctx, tConfig)
	case "AuthInfoStore":
		return coreAuth.NewDefaultAuthInfoStore(ctx, tConfig)
	case "PasswordChecker":
		return &audit.PasswordChecker{
			PwMinLength:            tConfig.UserAudit.PwMinLength,
			PwUppercaseRequired:    tConfig.UserAudit.PwUppercaseRequired,
			PwLowercaseRequired:    tConfig.UserAudit.PwLowercaseRequired,
			PwDigitRequired:        tConfig.UserAudit.PwDigitRequired,
			PwSymbolRequired:       tConfig.UserAudit.PwSymbolRequired,
			PwMinGuessableLevel:    tConfig.UserAudit.PwMinGuessableLevel,
			PwExcludedKeywords:     tConfig.UserAudit.PwExcludedKeywords,
			PwExcludedFields:       tConfig.UserAudit.PwExcludedFields,
			PwHistorySize:          tConfig.UserAudit.PwHistorySize,
			PwHistoryDays:          tConfig.UserAudit.PwHistoryDays,
			PasswordHistoryEnabled: tConfig.UserAudit.PwHistorySize > 0 || tConfig.UserAudit.PwHistoryDays > 0,
		}
	case "PasswordAuthProvider":
		// TODO:
		// from tConfig
		authRecordKeys := [][]string{[]string{"email"}, []string{"username"}}
		return password.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "provider_password", createLoggerMaskFormatter(tConfig)),
			authRecordKeys,
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "AnonymousAuthProvider":
		return anonymous.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "provider_anonymous", createLoggerMaskFormatter(tConfig)),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "CustomTokenAuthProvider":
		return customtoken.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "provider_custom_token", createLoggerMaskFormatter(tConfig)),
			tConfig.Auth.CustomTokenSecret,
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "HandlerLogger":
		return logging.CreateLoggerWithRequestID(requestID, "handler", createLoggerMaskFormatter(tConfig))
	case "UserProfileStore":
		switch tConfig.UserProfile.ImplName {
		default:
			panic("unrecgonized user profile store implementation: " + tConfig.UserProfile.ImplName)
		case "record":
			// use record based profile store
			roleStore := coreAuth.NewDefaultRoleStore(ctx, tConfig)
			recordStore := pq.NewSafeRecordStore(
				roleStore,
				// TODO: get from tconfig
				true,
				db.NewSQLBuilder("record", tConfig.AppName),
				db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
				logging.CreateLoggerWithRequestID(requestID, "record", createLoggerMaskFormatter(tConfig)),
				db.NewSafeTxContextWithContext(ctx, tConfig),
			)
			// TODO: get from tConfig
			assetStore := fs.NewAssetStore("", "", "", true, logging.CreateLoggerWithRequestID(requestID, "record", createLoggerMaskFormatter(tConfig)))
			return userprofile.NewUserProfileRecordStore(
				tConfig.UserProfile.ImplStoreURL,
				tConfig.APIKey,
				logging.CreateLoggerWithRequestID(requestID, "auth_user_profile", createLoggerMaskFormatter(tConfig)),
				coreAuth.NewContextGetterWithContext(ctx),
				db.NewTxContextWithContext(ctx, tConfig),
				recordStore,
				assetStore,
			)
		case "":
			// use auth default profile store
			return userprofile.NewSafeProvider(
				db.NewSQLBuilder("auth", tConfig.AppName),
				db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
				logging.CreateLoggerWithRequestID(requestID, "auth_user_profile", createLoggerMaskFormatter(tConfig)),
				db.NewSafeTxContextWithContext(ctx, tConfig),
			)
			// case "skygear":
			// 	return XXX
		}
	case "RoleStore":
		return coreAuth.NewDefaultRoleStore(ctx, tConfig)
	case "ForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "TestForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "ForgotPasswordCodeGenerator":
		return &forgotpwdemail.CodeGenerator{MasterKey: tConfig.MasterKey}
	case "ForgotPasswordSecureMatch":
		return tConfig.ForgotPassword.SecureMatch
	case "WelcomeEmailEnabled":
		return tConfig.WelcomeEmail.Enabled
	case "WelcomeEmailSendTask":
		return welcemail.NewSendTask(
			ctx,
			welcemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.SMTP)),
		)
	case "TestWelcomeEmailSender":
		return welcemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "IFrameHTMLProvider":
		return sso.NewIFrameHTMLProvider(tConfig.SSOSetting.URLPrefix, tConfig.SSOSetting.JSSDKCDNURL)
	case "VerifyCodeStore":
		return verifycode.NewStore(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "verify_code", createLoggerMaskFormatter(tConfig)),
		)
	case "UserVerifyCodeSenderFactory":
		return userverify.NewDefaultUserVerifyCodeSenderFactory(tConfig)
	case "AutoSendUserVerifyCodeOnSignup":
		return tConfig.UserVerify.AutoSendOnSignup
	case "UserVerifyKeys":
		return tConfig.UserVerify.Keys
	case "AuditTrail":
		trail, err := audit.NewTrail(tConfig.UserAudit.Enabled, tConfig.UserAudit.TrailHandlerURL)
		if err != nil {
			panic(err)
		}
		return trail
	case "SSOProviderFactory":
		return sso.NewProviderFactory(tConfig)
	case "AsyncTaskQueue":
		return async.NewQueue(ctx, requestID, tConfig, m.AsyncTaskExecutor)
	default:
		return nil
	}
}

func createLoggerMaskFormatter(tConfig config.TenantConfiguration) logrus.Formatter {
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}
