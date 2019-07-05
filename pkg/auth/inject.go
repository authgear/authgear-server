package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
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
			PwMinLength:         tConfig.UserConfig.UserAudit.Password.MinLength,
			PwUppercaseRequired: tConfig.UserConfig.UserAudit.Password.UppercaseRequired,
			PwLowercaseRequired: tConfig.UserConfig.UserAudit.Password.LowercaseRequired,
			PwDigitRequired:     tConfig.UserConfig.UserAudit.Password.DigitRequired,
			PwSymbolRequired:    tConfig.UserConfig.UserAudit.Password.SymbolRequired,
			PwMinGuessableLevel: tConfig.UserConfig.UserAudit.Password.MinimumGuessableLevel,
			PwExcludedKeywords:  tConfig.UserConfig.UserAudit.Password.ExcludedKeywords,
			//PwExcludedFields:       tConfig.UserConfig.UserAudit.Password.ExcludedFields,
			PwHistorySize:          tConfig.UserConfig.UserAudit.Password.HistorySize,
			PwHistoryDays:          tConfig.UserConfig.UserAudit.Password.HistoryDays,
			PasswordHistoryEnabled: tConfig.UserConfig.UserAudit.Password.HistorySize > 0 || tConfig.UserConfig.UserAudit.Password.HistoryDays > 0,
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
			tConfig.UserConfig.UserAudit.Password.HistorySize,
			tConfig.UserConfig.UserAudit.Password.HistoryDays,
			tConfig.UserConfig.UserAudit.Password.HistorySize > 0 || tConfig.UserConfig.UserAudit.Password.HistoryDays > 0,
		)
	case "PasswordAuthProvider":
		// TODO:
		// from tConfig
		passwordHistoryEnabled := tConfig.UserConfig.UserAudit.Password.HistorySize > 0 || tConfig.UserConfig.UserAudit.Password.HistoryDays > 0
		return password.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "provider_password", createLoggerMaskFormatter(tConfig)),
			tConfig.UserConfig.Auth.LoginIDKeys,
			tConfig.UserConfig.Auth.AllowedRealms,
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
			tConfig.UserConfig.Auth.CustomTokenSecret,
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "HandlerLogger":
		return logging.CreateLoggerWithRequestID(requestID, "handler", createLoggerMaskFormatter(tConfig))
	case "UserProfileStore":
		return userprofile.NewSafeProvider(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
			logging.CreateLoggerWithRequestID(requestID, "auth_user_profile", createLoggerMaskFormatter(tConfig)),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	case "ForgotPasswordEmailSender":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return forgotpwdemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.AppConfig.SMTP), templateEngine)
	case "TestForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.AppConfig.SMTP))
	case "ForgotPasswordCodeGenerator":
		return &forgotpwdemail.CodeGenerator{MasterKey: tConfig.UserConfig.MasterKey}
	case "ForgotPasswordSecureMatch":
		return tConfig.UserConfig.ForgotPassword.SecureMatch
	case "ResetPasswordHTMLProvider":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return forgotpwdemail.NewResetPasswordHTMLProvider(tConfig.UserConfig.ForgotPassword, templateEngine)
	case "WelcomeEmailEnabled":
		return tConfig.UserConfig.WelcomeEmail.Enabled
	case "WelcomeEmailDestination":
		return tConfig.UserConfig.WelcomeEmail.Destination
	case "WelcomeEmailSender":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return welcemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.AppConfig.SMTP), templateEngine)
	case "TestWelcomeEmailSender":
		return welcemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.AppConfig.SMTP))
	case "IFrameHTMLProvider":
		return sso.NewIFrameHTMLProvider(tConfig.UserConfig.SSO.APIEndpoint(), tConfig.UserConfig.SSO.JSSDKCDNURL)
	case "UserVerifyCodeSenderFactory":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return userverify.NewDefaultUserVerifyCodeSenderFactory(
			tConfig,
			templateEngine,
			logging.CreateLoggerWithRequestID(requestID, "code_sender", createLoggerMaskFormatter(tConfig)),
		)
	case "AutoSendUserVerifyCodeOnSignup":
		return tConfig.UserConfig.UserVerification.AutoSendOnSignupDisabled
	case "UserVerifyKeys":
		return tConfig.UserConfig.UserVerification.LoginIDKeys
	case "UserVerificationProvider":
		return userverify.NewProvider(
			userverify.NewCodeGenerator(tConfig),
			userverify.NewSafeStore(
				db.NewSQLBuilder("auth", tConfig.AppName),
				db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig)),
				logging.CreateLoggerWithRequestID(requestID, "verify_code", createLoggerMaskFormatter(tConfig)),
				db.NewSafeTxContextWithContext(ctx, tConfig),
			),
			tConfig.UserConfig.UserVerification,
			time.NewProvider(),
		)
	case "VerifyHTMLProvider":
		templateEngine := authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
		return userverify.NewVerifyHTMLProvider(tConfig.UserConfig.UserVerification, templateEngine)
	case "AuditTrail":
		trail, err := audit.NewTrail(tConfig.UserConfig.UserAudit.Enabled, tConfig.UserConfig.UserAudit.TrailHandlerURL)
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
		return sso.NewAuthHandlerHTMLProvider(tConfig.UserConfig.SSO.APIEndpoint(), tConfig.UserConfig.SSO.JSSDKCDNURL)
	case "AsyncTaskQueue":
		return async.NewQueue(ctx, requestID, tConfig, m.AsyncTaskExecutor)
	case "HookStore":
		l := logging.CreateLoggerWithRequestID(requestID, "auth_hook", createLoggerMaskFormatter(tConfig))
		return hook.NewHookProvider(tConfig.Hooks, hook.ExecutorImpl{}, l, requestID)
	default:
		return nil
	}
}

func createLoggerMaskFormatter(tConfig config.TenantConfiguration) logrus.Formatter {
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}
