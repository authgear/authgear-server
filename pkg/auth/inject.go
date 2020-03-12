package auth

import (
	"context"
	"net/http"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	mfaPQ "github.com/skygeario/skygear-server/pkg/auth/dependency/mfa/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	pqPWHistory "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"
	authTemplate "github.com/skygeario/skygear-server/pkg/auth/template"
	"github.com/skygeario/skygear-server/pkg/core/apiclientconfig"
	"github.com/skygeario/skygear-server/pkg/core/async"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	pqAuthInfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	redisSession "github.com/skygeario/skygear-server/pkg/core/auth/session/redis"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sentry"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/template"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type DependencyMap struct {
	EnableFileSystemTemplate bool
	Validator                *validation.Validator
	AssetGearLoader          *template.AssetGearLoader
	AsyncTaskExecutor        *async.Executor
	UseInsecureCookie        bool
	DefaultConfiguration     config.DefaultConfiguration
	ReservedNameChecker      *loginid.ReservedNameChecker
}

// Provide provides dependency instance by name
// nolint: gocyclo, golint
func (m DependencyMap) Provide(
	dependencyName string,
	request *http.Request,
	ctx context.Context,
	requestID string,
	tc config.TenantConfiguration,
) interface{} {
	// populate default
	appConfig := tc.AppConfig
	if !appConfig.SMTP.IsValid() {
		appConfig.SMTP = m.DefaultConfiguration.SMTP
	}

	if !appConfig.Twilio.IsValid() {
		appConfig.Twilio = m.DefaultConfiguration.Twilio
	}

	if !appConfig.Nexmo.IsValid() {
		appConfig.Nexmo = m.DefaultConfiguration.Nexmo
	}

	// To avoid mutating tc
	tConfig := tc
	tConfig.AppConfig = appConfig

	newLoggerFactory := func() logging.Factory {
		logHook := logging.NewDefaultLogHook(tConfig.DefaultSensitiveLoggerValues())
		sentryHook := sentry.NewLogHookFromContext(ctx)
		if request == nil {
			return logging.NewFactoryFromRequestID(requestID, logHook, sentryHook)
		} else {
			return logging.NewFactoryFromRequest(request, logHook, sentryHook)
		}
	}

	newSQLBuilder := func() db.SQLBuilder {
		return db.NewSQLBuilder("auth", tConfig.DatabaseConfig.DatabaseSchema, tConfig.AppID)
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

	newPasswordStore := func() password.Store {
		return password.NewStore(
			newSQLBuilder(),
			newSQLExecutor(),
		)
	}

	newPasswordHistoryStore := func() passwordhistory.Store {
		return pqPWHistory.NewPasswordHistoryStore(
			newTimeProvider(),
			newSQLBuilder(),
			newSQLExecutor(),
		)
	}

	newTemplateEngine := func() *template.Engine {
		return authTemplate.NewEngineWithConfig(
			tConfig,
			m.EnableFileSystemTemplate,
			m.AssetGearLoader,
		)
	}

	newAuthInfoStore := func() authinfo.Store {
		return pqAuthInfo.NewAuthInfoStore(
			db.NewSQLBuilder("core", tConfig.DatabaseConfig.DatabaseSchema, tConfig.AppID),
			newSQLExecutor(),
		)
	}

	newUserProfileStore := func() userprofile.Store {
		return userprofile.NewUserProfileStore(
			newTimeProvider(),
			newSQLBuilder(),
			newSQLExecutor(),
		)
	}

	// TODO:
	// from tConfig
	isPasswordHistoryEnabled := func() bool {
		return tConfig.AppConfig.PasswordPolicy.HistorySize > 0 ||
			tConfig.AppConfig.PasswordPolicy.HistoryDays > 0
	}

	newPasswordAuthProvider := func() password.Provider {
		return password.NewProvider(
			newTimeProvider(),
			newPasswordStore(),
			newPasswordHistoryStore(),
			newLoggerFactory(),
			tConfig.AppConfig.Auth.LoginIDKeys,
			tConfig.AppConfig.Auth.LoginIDTypes,
			tConfig.AppConfig.Auth.AllowedRealms,
			isPasswordHistoryEnabled(),
			m.ReservedNameChecker,
		)
	}

	newOAuthAuthProvider := func() oauth.Provider {
		return oauth.NewProvider(
			newSQLBuilder(),
			newSQLExecutor(),
		)
	}

	newHookProvider := func() hook.Provider {
		return inject.Scoped(ctx, "HookProvider", func() interface{} {
			return hook.NewProvider(
				requestID,
				urlprefix.NewProvider(request),
				hook.NewStore(newSQLBuilder(), newSQLExecutor()),
				newAuthContext(),
				newTimeProvider(),
				newAuthInfoStore(),
				newUserProfileStore(),
				hook.NewDeliverer(
					&tConfig,
					newTimeProvider(),
					hook.NewMutator(
						tConfig.AppConfig.UserVerification,
						newPasswordAuthProvider(),
						newAuthInfoStore(),
						newUserProfileStore(),
					),
				),
				newLoggerFactory(),
			)
		})().(hook.Provider)
	}

	newSessionProvider := func() session.Provider {
		return session.NewProvider(
			request,
			redisSession.NewStore(ctx, tConfig.AppID, newTimeProvider(), newLoggerFactory()),
			redisSession.NewEventStore(ctx, tConfig.AppID),
			newAuthContext(),
			tConfig.AppConfig.Clients,
		)
	}

	newIdentityProvider := func() principal.IdentityProvider {
		return principal.NewIdentityProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			newOAuthAuthProvider(),
			newPasswordAuthProvider(),
		)
	}

	newSessionWriter := func() session.Writer {
		return session.NewWriter(
			newAuthContext(),
			tConfig.AppConfig.Clients,
			tConfig.AppConfig.MFA,
			m.UseInsecureCookie,
		)
	}

	newSMSClient := func() sms.Client {
		return sms.NewClient(tConfig.AppConfig)
	}

	newMailSender := func() mail.Sender {
		return mail.NewSender(tConfig.AppConfig.SMTP)
	}

	newMFAProvider := func() mfa.Provider {
		return mfa.NewProvider(
			mfaPQ.NewStore(
				tConfig.AppConfig.MFA,
				newSQLBuilder(),
				newSQLExecutor(),
				newTimeProvider(),
			),
			tConfig.AppConfig.MFA,
			newTimeProvider(),
			mfa.NewSender(
				tConfig,
				newSMSClient(),
				newMailSender(),
				newTemplateEngine(),
			),
		)
	}

	newLoginIDNormalizerFactory := func() loginid.LoginIDNormalizerFactory {
		return loginid.NewLoginIDNormalizerFactory(
			tConfig.AppConfig.Auth.LoginIDKeys,
			tConfig.AppConfig.Auth.LoginIDTypes,
		)
	}

	newOAuthUserInfoDecoder := func() sso.UserInfoDecoder {
		return sso.NewUserInfoDecoder(newLoginIDNormalizerFactory())
	}

	newAPIClientConfigurationProvider := func() apiclientconfig.Provider {
		return apiclientconfig.NewProvider(newAuthContext(), tConfig)
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
	case "RequireAuthz":
		return handler.NewRequireAuthzFactory(newAuthContext(), newLoggerFactory())
	case "Validator":
		return m.Validator
	case "SessionProvider":
		return newSessionProvider()
	case "SessionWriter":
		return newSessionWriter()
	case "MFAProvider":
		return newMFAProvider()
	case "AuthnSessionProvider":
		return authnsession.NewProvider(
			newAuthContext(),
			tConfig.AppConfig.MFA,
			tConfig.AppConfig.Auth.AuthenticationSession,
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
			PwMinLength:         tConfig.AppConfig.PasswordPolicy.MinLength,
			PwUppercaseRequired: tConfig.AppConfig.PasswordPolicy.UppercaseRequired,
			PwLowercaseRequired: tConfig.AppConfig.PasswordPolicy.LowercaseRequired,
			PwDigitRequired:     tConfig.AppConfig.PasswordPolicy.DigitRequired,
			PwSymbolRequired:    tConfig.AppConfig.PasswordPolicy.SymbolRequired,
			PwMinGuessableLevel: tConfig.AppConfig.PasswordPolicy.MinimumGuessableLevel,
			PwExcludedKeywords:  tConfig.AppConfig.PasswordPolicy.ExcludedKeywords,
			//PwExcludedFields:       tConfig.AppConfig.PasswordPolicy.ExcludedFields,
			PwHistorySize:          tConfig.AppConfig.PasswordPolicy.HistorySize,
			PwHistoryDays:          tConfig.AppConfig.PasswordPolicy.HistoryDays,
			PasswordHistoryEnabled: tConfig.AppConfig.PasswordPolicy.HistorySize > 0 || tConfig.AppConfig.PasswordPolicy.HistoryDays > 0,
			PasswordHistoryStore:   newPasswordHistoryStore(),
		}
	case "PwHousekeeper":
		return authAudit.NewPwHousekeeper(
			newPasswordHistoryStore(),
			newLoggerFactory(),
			tConfig.AppConfig.PasswordPolicy.HistorySize,
			tConfig.AppConfig.PasswordPolicy.HistoryDays,
			isPasswordHistoryEnabled(),
		)
	case "PasswordAuthProvider":
		return newPasswordAuthProvider()
	case "HandlerLogger":
		return newLoggerFactory().NewLogger("handler")
	case "UserProfileStore":
		return newUserProfileStore()
	case "ForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultSender(tConfig, urlprefix.NewProvider(request).Value(), newMailSender(), newTemplateEngine())
	case "ForgotPasswordCodeGenerator":
		return &forgotpwdemail.CodeGenerator{MasterKey: tConfig.AppConfig.MasterKey}
	case "ForgotPasswordSecureMatch":
		return tConfig.AppConfig.ForgotPassword.SecureMatch
	case "ResetPasswordHTMLProvider":
		return forgotpwdemail.NewResetPasswordHTMLProvider(urlprefix.NewProvider(request).Value(), tConfig.AppConfig.ForgotPassword, newTemplateEngine())
	case "WelcomeEmailEnabled":
		return tConfig.AppConfig.WelcomeEmail.Enabled
	case "WelcomeEmailDestination":
		return tConfig.AppConfig.WelcomeEmail.Destination
	case "WelcomeEmailSender":
		return welcemail.NewDefaultSender(tConfig, newMailSender(), newTemplateEngine())
	case "UserVerifyCodeSenderFactory":
		return userverify.NewDefaultUserVerifyCodeSenderFactory(
			tConfig,
			newTemplateEngine(),
			newMailSender(),
			newSMSClient(),
		)
	case "AutoSendUserVerifyCodeOnSignup":
		return tConfig.AppConfig.UserVerification.AutoSendOnSignup
	case "UserVerifyLoginIDKeys":
		return tConfig.AppConfig.UserVerification.LoginIDKeys
	case "UserVerificationProvider":
		return userverify.NewProvider(
			userverify.NewCodeGenerator(tConfig),
			userverify.NewStore(
				newSQLBuilder(),
				newSQLExecutor(),
			),
			tConfig.AppConfig.UserVerification,
			newTimeProvider(),
		)
	case "VerifyHTMLProvider":
		return userverify.NewVerifyHTMLProvider(tConfig.AppConfig.UserVerification, newTemplateEngine())
	case "LoginIDNormalizerFactory":
		return newLoginIDNormalizerFactory()
	case "OAuthUserInfoDecoder":
		return newOAuthUserInfoDecoder()
	case "SSOOAuthProviderFactory":
		return sso.NewOAuthProviderFactory(tConfig, urlprefix.NewProvider(request), newTimeProvider(), newOAuthUserInfoDecoder(), newLoginIDNormalizerFactory())
	case "SSOProvider":
		return sso.NewProvider(
			tConfig.AppID,
			newAPIClientConfigurationProvider(),
			tConfig.AppConfig.SSO.OAuth,
		)
	case "OAuthAuthProvider":
		return newOAuthAuthProvider()
	case "IdentityProvider":
		return newIdentityProvider()
	case "AuthHandlerHTMLProvider":
		return sso.NewAuthHandlerHTMLProvider(urlprefix.NewProvider(request).Value())
	case "AsyncTaskQueue":
		return async.NewQueue(ctx, requestID, tConfig, m.AsyncTaskExecutor)
	case "HookProvider":
		return newHookProvider()
	case "OAuthConfiguration":
		return tConfig.AppConfig.SSO.OAuth
	case "AuthConfiguration":
		return *tConfig.AppConfig.Auth
	case "MFAConfiguration":
		return *tConfig.AppConfig.MFA
	case "APIClientConfigurationProvider":
		return newAPIClientConfigurationProvider()
	case "URLPrefix":
		return urlprefix.NewProvider(request).Value()
	case "TemplateEngine":
		return newTemplateEngine()
	case "TimeProvider":
		return newTimeProvider()
	default:
		return nil
	}
}
