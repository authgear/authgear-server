package auth

import (
	"context"
	"net/http"

	authAudit "github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authnsession"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	mfaPQ "github.com/skygeario/skygear-server/pkg/auth/dependency/mfa/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	pqPWHistory "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/customtoken"
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
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	pqAuthInfo "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	redisSession "github.com/skygeario/skygear-server/pkg/core/auth/session/redis"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
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
	userConfig := tc.UserConfig
	if !userConfig.SMTP.IsValid() {
		userConfig.SMTP = m.DefaultConfiguration.SMTP
	}

	if !userConfig.Twilio.IsValid() {
		userConfig.Twilio = m.DefaultConfiguration.Twilio
	}

	if !userConfig.Nexmo.IsValid() {
		userConfig.Nexmo = m.DefaultConfiguration.Nexmo
	}

	// To avoid mutating tc
	tConfig := tc
	tConfig.UserConfig = userConfig

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
			newLoggerFactory(),
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
			db.NewSQLBuilder("core", tConfig.AppConfig.DatabaseSchema, tConfig.AppID),
			newSQLExecutor(),
		)
	}

	newUserProfileStore := func() userprofile.Store {
		return userprofile.NewUserProfileStore(
			newSQLBuilder(),
			newSQLExecutor(),
		)
	}

	// TODO:
	// from tConfig
	isPasswordHistoryEnabled := func() bool {
		return tConfig.UserConfig.PasswordPolicy.HistorySize > 0 ||
			tConfig.UserConfig.PasswordPolicy.HistoryDays > 0
	}

	newPasswordAuthProvider := func() password.Provider {
		return password.NewProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			newLoggerFactory(),
			tConfig.UserConfig.Auth.LoginIDKeys,
			tConfig.UserConfig.Auth.AllowedRealms,
			isPasswordHistoryEnabled(),
		)
	}

	newCustomTokenAuthProvider := func() customtoken.Provider {
		return customtoken.NewProvider(
			newSQLBuilder(),
			newSQLExecutor(),
			tConfig.UserConfig.SSO.CustomToken,
		)
	}

	newOAuthAuthProvider := func() oauth.Provider {
		return oauth.NewProvider(
			newSQLBuilder(),
			newSQLExecutor(),
		)
	}

	newHookProvider := func() hook.Provider {
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
					tConfig.UserConfig.UserVerification,
					newPasswordAuthProvider(),
					newAuthInfoStore(),
					newUserProfileStore(),
				),
			),
			newLoggerFactory(),
		)
	}

	newSessionProvider := func() session.Provider {
		return session.NewProvider(
			request,
			redisSession.NewStore(ctx, tConfig.AppID, newTimeProvider(), newLoggerFactory()),
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
		return sms.NewClient(tConfig.UserConfig)
	}

	newMailSender := func() mail.Sender {
		return mail.NewSender(tConfig.UserConfig.SMTP)
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
				newMailSender(),
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
			PwMinLength:         tConfig.UserConfig.PasswordPolicy.MinLength,
			PwUppercaseRequired: tConfig.UserConfig.PasswordPolicy.UppercaseRequired,
			PwLowercaseRequired: tConfig.UserConfig.PasswordPolicy.LowercaseRequired,
			PwDigitRequired:     tConfig.UserConfig.PasswordPolicy.DigitRequired,
			PwSymbolRequired:    tConfig.UserConfig.PasswordPolicy.SymbolRequired,
			PwMinGuessableLevel: tConfig.UserConfig.PasswordPolicy.MinimumGuessableLevel,
			PwExcludedKeywords:  tConfig.UserConfig.PasswordPolicy.ExcludedKeywords,
			//PwExcludedFields:       tConfig.UserConfig.PasswordPolicy.ExcludedFields,
			PwHistorySize:          tConfig.UserConfig.PasswordPolicy.HistorySize,
			PwHistoryDays:          tConfig.UserConfig.PasswordPolicy.HistoryDays,
			PasswordHistoryEnabled: tConfig.UserConfig.PasswordPolicy.HistorySize > 0 || tConfig.UserConfig.PasswordPolicy.HistoryDays > 0,
			PasswordHistoryStore:   newPasswordHistoryStore(),
		}
	case "PwHousekeeper":
		return authAudit.NewPwHousekeeper(
			newPasswordHistoryStore(),
			newLoggerFactory(),
			tConfig.UserConfig.PasswordPolicy.HistorySize,
			tConfig.UserConfig.PasswordPolicy.HistoryDays,
			isPasswordHistoryEnabled(),
		)
	case "PasswordAuthProvider":
		return newPasswordAuthProvider()
	case "CustomTokenAuthProvider":
		return newCustomTokenAuthProvider()
	case "HandlerLogger":
		return newLoggerFactory().NewLogger("handler")
	case "UserProfileStore":
		return newUserProfileStore()
	case "ForgotPasswordEmailSender":
		return forgotpwdemail.NewDefaultSender(tConfig, urlprefix.NewProvider(request).Value(), newMailSender(), newTemplateEngine())
	case "ForgotPasswordCodeGenerator":
		return &forgotpwdemail.CodeGenerator{MasterKey: tConfig.UserConfig.MasterKey}
	case "ForgotPasswordSecureMatch":
		return tConfig.UserConfig.ForgotPassword.SecureMatch
	case "ResetPasswordHTMLProvider":
		return forgotpwdemail.NewResetPasswordHTMLProvider(urlprefix.NewProvider(request).Value(), tConfig.UserConfig.ForgotPassword, newTemplateEngine())
	case "WelcomeEmailEnabled":
		return tConfig.UserConfig.WelcomeEmail.Enabled
	case "WelcomeEmailDestination":
		return tConfig.UserConfig.WelcomeEmail.Destination
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
		return tConfig.UserConfig.UserVerification.AutoSendOnSignup
	case "UserVerifyLoginIDKeys":
		return tConfig.UserConfig.UserVerification.LoginIDKeys
	case "UserVerificationProvider":
		return userverify.NewProvider(
			userverify.NewCodeGenerator(tConfig),
			userverify.NewStore(
				newSQLBuilder(),
				newSQLExecutor(),
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
	case "SSOOAuthProviderFactory":
		return sso.NewOAuthProviderFactory(tConfig, urlprefix.NewProvider(request))
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
	case "CustomTokenConfiguration":
		return tConfig.UserConfig.SSO.CustomToken
	case "OAuthConfiguration":
		return tConfig.UserConfig.SSO.OAuth
	case "AuthConfiguration":
		return *tConfig.UserConfig.Auth
	case "MFAConfiguration":
		return *tConfig.UserConfig.MFA
	case "APIClientConfigurationProvider":
		return apiclientconfig.NewProvider(newAuthContext(), tConfig)
	case "URLPrefix":
		return urlprefix.NewProvider(request).Value()
	case "TemplateEngine":
		return newTemplateEngine()
	default:
		return nil
	}
}
