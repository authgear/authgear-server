package auth

import (
	"context"
	"net/http"

	"github.com/go-gomail/gomail"
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	mfaPQ "github.com/skygeario/skygear-server/pkg/auth/dependency/mfa/pq"
	pqPWHistory "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/customtoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	redisSession "github.com/skygeario/skygear-server/pkg/core/auth/session/redis"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

type DependencyMap struct {
	TemplateEngine    *template.Engine
	AsyncTaskExecutor *async.Executor
	UseInsecureCookie bool
}

// Provide provides dependency instance by name
// nolint: gocyclo, golint
func (m DependencyMap) Provide(
	dependencyName string,
	request *http.Request,
	ctx context.Context,
	requestID string,
	tConfig config.TenantConfiguration,
) interface{} {
	newLogger := func(name string) *logrus.Entry {
		formatter := logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
		return logging.CreateLoggerWithRequestID(requestID, name, formatter)
	}

	newLoggerFactory := func() logging.Factory {
		formatter := logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
		return logging.NewFactory(request, formatter)
	}

	newSQLBuilder := func() db.SQLBuilder {
		return db.NewSQLBuilder("auth", tConfig.AppConfig.DatabaseSchema, tConfig.AppID)
	}

	newSQLExecutor := func() db.SQLExecutor {
		return db.NewSQLExecutor(ctx, db.NewContextWithContext(ctx, tConfig))
	}

	newTimeProvider := func() time.Provider {
		return time.NewProvider()
	}

	newAuthContext := func() coreAuth.ContextGetter {
		return coreAuth.NewContextGetterWithContext(ctx)
	}

	newPasswordHistoryStore := func() passwordhistory.Store {
		return pqPWHistory.NewPasswordHistoryStore(
			newSQLBuilder(),
			newSQLExecutor(),
			newLogger("auth_password_history"),
		)
	}

	newTemplateEngine := func() *template.Engine {
		return authTemplate.NewEngineWithConfig(m.TemplateEngine, tConfig)
	}

	newAuthInfoStore := func() authinfo.Store {
		return coreAuth.NewDefaultAuthInfoStore(ctx, tConfig)
	}

	newUserProfileStore := func() userprofile.Store {
		return userprofile.NewSafeProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			newLogger("auth_user_profile"),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	}

	// TODO:
	// from tConfig
	isPasswordHistoryEnabled := func() bool {
		return tConfig.UserConfig.UserAudit.Password.HistorySize > 0 ||
			tConfig.UserConfig.UserAudit.Password.HistoryDays > 0
	}

	newPasswordAuthProvider := func() password.Provider {
		return password.NewSafeProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			newLogger("provider_password"),
			tConfig.UserConfig.Auth.LoginIDKeys,
			tConfig.UserConfig.Auth.AllowedRealms,
			isPasswordHistoryEnabled(),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	}

	newCustomTokenAuthProvider := func() customtoken.Provider {
		return customtoken.NewSafeProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			newLogger("provider_custom_token"),
			tConfig.UserConfig.SSO.CustomToken,
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	}

	newOAuthAuthProvider := func() oauth.Provider {
		return oauth.NewSafeProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			newLogger("provider_oauth"),
			db.NewSafeTxContextWithContext(ctx, tConfig),
		)
	}

	newHookProvider := func() hook.Provider {
		return hook.NewProvider(
			requestID,
			request,
			hook.NewStore(newSQLBuilder(), newSQLExecutor()),
			newAuthContext(),
			newTimeProvider(),
			newAuthInfoStore(),
			newUserProfileStore(),
			hook.NewDeliverer(
				&tConfig,
				newTimeProvider(),
				hook.NewMutator(
					&tConfig.UserConfig.UserVerification,
					newPasswordAuthProvider(),
					newAuthInfoStore(),
					newUserProfileStore(),
				),
			),
		)
	}

	newSessionProvider := func() session.Provider {
		return session.NewProvider(
			request,
			redisSession.NewStore(ctx, tConfig.AppID, newTimeProvider()),
			redisSession.NewEventStore(ctx, tConfig.AppID),
			newAuthContext(),
			tConfig.UserConfig.Clients,
		)
	}

	newIdentityProvider := func() principal.IdentityProvider {
		return principal.NewIdentityProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			newCustomTokenAuthProvider(),
			newOAuthAuthProvider(),
			newPasswordAuthProvider(),
		)
	}

	newSessionWriter := func() session.Writer {
		return session.NewWriter(
			newAuthContext(),
			tConfig.UserConfig.Clients,
			tConfig.UserConfig.MFA,
			m.UseInsecureCookie,
		)
	}

	newSMSClient := func() sms.Client {
		return sms.NewClient(tConfig.AppConfig)
	}

	newMailDialer := func() *gomail.Dialer {
		return mail.NewDialer(tConfig.AppConfig.SMTP)
	}

	newMFAProvider := func() mfa.Provider {
		return mfa.NewProvider(
			mfaPQ.NewStore(
				tConfig.UserConfig.MFA,
				newSQLBuilder(),
				newSQLExecutor(),
				newTimeProvider(),
			),
			tConfig.UserConfig.MFA,
			newTimeProvider(),
			mfa.NewSender(
				newSMSClient(),
				newMailDialer(),
				newTemplateEngine(),
			),
		)
	}

	switch dependencyName {
	case "AuthContextGetter":
		return newAuthContext()
	case "AuthContextSetter":
		return coreAuth.NewContextSetterWithContext(ctx)
	case "TxContext":
		return db.NewTxContextWithContext(ctx, tConfig)
	case "LoggerFactory":
		return newLoggerFactory()
	case "SessionProvider":
		return newSessionProvider()
	case "SessionWriter":
		return newSessionWriter()
	case "MFAProvider":
		return newMFAProvider()
	case "AuthnSessionProvider":
		return authnsession.NewProvider(
			newAuthContext(),
			tConfig.UserConfig.MFA,
			tConfig.UserConfig.Auth.AuthenticationSession,
			newTimeProvider(),
			newMFAProvider(),
			newAuthInfoStore(),
			newSessionProvider(),
			newSessionWriter(),
			newIdentityProvider(),
			newHookProvider(),
			newUserProfileStore(),
		)
	case "AuthInfoStore":
		return newAuthInfoStore()
	case "PasswordChecker":
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
			PasswordHistoryStore:   newPasswordHistoryStore(),
		}
	case "PwHousekeeper":
		return authAudit.NewPwHousekeeper(
			newPasswordHistoryStore(),
			newLogger("audit"),
			tConfig.UserConfig.UserAudit.Password.HistorySize,
			tConfig.UserConfig.UserAudit.Password.HistoryDays,
			isPasswordHistoryEnabled(),
		)
	case "PasswordAuthProvider":
		return newPasswordAuthProvider()
	case "CustomTokenAuthProvider":
		return newCustomTokenAuthProvider()
	case "HandlerLogger":
		return newLogger("handler")
	case "UserProfileStore":
		return newUserProfileStore()
	case "ForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultSender(tConfig, newMailDialer(), newTemplateEngine())
	case "TestForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultTestSender(tConfig, newMailDialer())
	case "ForgotPasswordCodeGenerator":
		return &forgotpwdemail.CodeGenerator{MasterKey: tConfig.UserConfig.MasterKey}
	case "ForgotPasswordSecureMatch":
		return tConfig.UserConfig.ForgotPassword.SecureMatch
	case "ResetPasswordHTMLProvider":
		return forgotpwdemail.NewResetPasswordHTMLProvider(tConfig.UserConfig.ForgotPassword, newTemplateEngine())
	case "WelcomeEmailEnabled":
		return tConfig.UserConfig.WelcomeEmail.Enabled
	case "WelcomeEmailDestination":
		return tConfig.UserConfig.WelcomeEmail.Destination
	case "WelcomeEmailSender":
		return welcemail.NewDefaultSender(tConfig, newMailDialer(), newTemplateEngine())
	case "TestWelcomeEmailSender":
		return welcemail.NewDefaultTestSender(tConfig, newMailDialer())
	case "IFrameHTMLProvider":
		return sso.NewIFrameHTMLProvider(tConfig.UserConfig.SSO.OAuth.APIEndpoint())
	case "UserVerifyCodeSenderFactory":
		return userverify.NewDefaultUserVerifyCodeSenderFactory(tConfig, newTemplateEngine())
	case "UserVerifyTestCodeSenderFactory":
		return userverify.NewDefaultUserVerifyTestCodeSenderFactory(tConfig, newTemplateEngine())
	case "AutoSendUserVerifyCodeOnSignup":
		return tConfig.UserConfig.UserVerification.AutoSendOnSignup
	case "UserVerifyLoginIDKeys":
		return tConfig.UserConfig.UserVerification.LoginIDKeys
	case "UserVerificationProvider":
		return userverify.NewProvider(
			userverify.NewCodeGenerator(tConfig),
			userverify.NewSafeStore(
				newSQLBuilder(),
				newSQLExecutor(),
				newLogger("verify_code"),
				db.NewSafeTxContextWithContext(ctx, tConfig),
			),
			tConfig.UserConfig.UserVerification,
			newTimeProvider(),
		)
	case "VerifyHTMLProvider":
		return userverify.NewVerifyHTMLProvider(tConfig.UserConfig.UserVerification, newTemplateEngine())
	case "AuditTrail":
		trail, err := audit.NewTrail(tConfig.UserConfig.UserAudit.Enabled, tConfig.UserConfig.UserAudit.TrailHandlerURL)
		if err != nil {
			panic(err)
		}
		return trail
	case "SSOProviderFactory":
		return sso.NewProviderFactory(tConfig)
	case "OAuthAuthProvider":
		return newOAuthAuthProvider()
	case "IdentityProvider":
		return newIdentityProvider()
	case "AuthHandlerHTMLProvider":
		return sso.NewAuthHandlerHTMLProvider(tConfig.UserConfig.SSO.OAuth.APIEndpoint())
	case "AsyncTaskQueue":
		return async.NewQueue(ctx, requestID, tConfig, m.AsyncTaskExecutor)
	case "HookProvider":
		return newHookProvider()
	case "CustomTokenConfiguration":
		return tConfig.UserConfig.SSO.CustomToken
	case "OAuthConfiguration":
		return tConfig.UserConfig.SSO.OAuth
	case "AuthConfiguration":
		return tConfig.UserConfig.Auth
	case "MFAConfiguration":
		return tConfig.UserConfig.MFA
	default:
		return nil
	}
}
