package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"

	"github.com/sirupsen/logrus"
	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	pqPWHistory "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/customtoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type DependencyMap struct {
	TemplateEngine    *template.Engine
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
		passwordHistoryStore := pqPWHistory.NewPasswordHistoryStore(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "auth_password_history", createLoggerMaskFormatter(tConfig)),
		)
		return &authAudit.PasswordChecker{
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
			PasswordHistoryStore:   passwordHistoryStore,
		}
	case "PwHousekeeper":
		passwordHistoryStore := pqPWHistory.NewPasswordHistoryStore(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "auth_password_history", createLoggerMaskFormatter(tConfig)),
		)
		return authAudit.NewPwHousekeeper(
			passwordHistoryStore,
			logging.CreateLoggerWithRequestID(requestID, "audit", createLoggerMaskFormatter(tConfig)),
			tConfig.UserAudit.PwHistorySize,
			tConfig.UserAudit.PwHistoryDays,
			tConfig.UserAudit.PwHistorySize > 0 || tConfig.UserAudit.PwHistoryDays > 0,
		)
	case "PasswordAuthProvider":
		// TODO:
		// from tConfig
		passwordHistoryEnabled := tConfig.UserAudit.PwHistorySize > 0 || tConfig.UserAudit.PwHistoryDays > 0
		return password.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "provider_password", createLoggerMaskFormatter(tConfig)),
			tConfig.Auth.AuthRecordKeys,
			passwordHistoryEnabled,
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
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return forgotpwdemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.SMTP), templateEngine)
	case "TestForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "ForgotPasswordCodeGenerator":
		return &forgotpwdemail.CodeGenerator{MasterKey: tConfig.MasterKey}
	case "ForgotPasswordSecureMatch":
		return tConfig.ForgotPassword.SecureMatch
	case "ResetPasswordHTMLProvider":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return forgotpwdemail.NewResetPasswordHTMLProvider(tConfig.ForgotPassword, templateEngine)
	case "WelcomeEmailEnabled":
		return tConfig.WelcomeEmail.Enabled
	case "WelcomeEmailSender":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return welcemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.SMTP), templateEngine)
	case "TestWelcomeEmailSender":
		return welcemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "IFrameHTMLProvider":
		return sso.NewIFrameHTMLProvider(tConfig.SSOSetting.APIEndpoint(), tConfig.SSOSetting.JSSDKCDNURL)
	case "VerifyCodeStore":
		return userverify.NewSafeStore(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "verify_code", createLoggerMaskFormatter(tConfig)),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "VerifyCodeCodeGeneratorFactory":
		return userverify.NewDefaultCodeGeneratorFactory(tConfig)
	case "UserVerifyCodeSenderFactory":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return userverify.NewDefaultUserVerifyCodeSenderFactory(tConfig, templateEngine)
	case "UserVerifyTestCodeSenderFactory":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return userverify.NewDefaultUserVerifyTestCodeSenderFactory(tConfig, templateEngine)
	case "AutoSendUserVerifyCodeOnSignup":
		return tConfig.UserVerify.AutoSendOnSignup
	case "UserVerifyKeys":
		return tConfig.UserVerify.Keys
	case "AutoUpdateUserVerifyFunc":
		return userverify.CreateAutoUpdateUserVerifyfunc(tConfig)
	case "AutoUpdateUserVerified":
		return tConfig.UserVerify.AutoUpdate
	case "VerifyHTMLProvider":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return userverify.NewVerifyHTMLProvider(tConfig.UserVerify, templateEngine)
	case "AuditTrail":
		trail, err := audit.NewTrail(tConfig.UserAudit.Enabled, tConfig.UserAudit.TrailHandlerURL)
		if err != nil {
			panic(err)
		}
		return trail
	case "SSOProviderFactory":
		return sso.NewProviderFactory(tConfig)
	case "OAuthAuthProvider":
		return oauth.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "provider_oauth", createLoggerMaskFormatter(tConfig)),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "AuthHandlerHTMLProvider":
		return sso.NewAuthHandlerHTMLProvider(tConfig.SSOSetting.APIEndpoint(), tConfig.SSOSetting.JSSDKCDNURL)
	case "AsyncTaskQueue":
		return async.NewQueue(ctx, requestID, tConfig, m.AsyncTaskExecutor)
	default:
		return nil
	}
}

func createLoggerMaskFormatter(tConfig config.TenantConfiguration) logrus.Formatter {
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}
