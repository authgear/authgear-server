package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/forgotpwdemail"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userverify/verifycode"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/welcemail"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/provider/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/asset/fs"
	"github.com/skygeario/skygear-server/pkg/core/audit"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record/pq" // tolerant nextimportslint: record
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
		tConfig := config.GetTenantConfig(r)
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
		trail, err := audit.NewTrail(tConfig.UserAudit.Enabled, tConfig.UserAudit.TrailHandlerURL, r)
		if err != nil {
			panic(err)
		}
		return trail
	case "ForgotPasswordEmailSender":
		tConfig := config.GetTenantConfig(r)
		return forgotpwdemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "TestForgotPasswordEmailSender":
		tConfig := config.GetTenantConfig(r)
		return forgotpwdemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "ForgotPasswordSecureMatch":
		tConfig := config.GetTenantConfig(r)
		return tConfig.ForgotPassword.SecureMatch
	case "WelcomeEmailSendTask":
		tConfig := config.GetTenantConfig(r)
		if !tConfig.WelcomeEmail.Enabled {
			return nil
		}

		task := welcemail.NewSendTask(welcemail.NewDefaultSender(tConfig, mail.NewDialer(tConfig.SMTP)))
		task.WaitForRequest(r.Context())
		return task
	case "TestWelcomeEmailSender":
		tConfig := config.GetTenantConfig(r)
		return welcemail.NewDefaultTestSender(tConfig, mail.NewDialer(tConfig.SMTP))
	case "SSOProvider":
		vars := mux.Vars(r)
		providerName := vars["provider"]
		tConfig := config.GetTenantConfig(r)
		SSOConf := tConfig.GetSSOConfigByName(providerName)
		SSOSetting := tConfig.SSOSetting
		setting := sso.Setting{
			URLPrefix:            SSOSetting.URLPrefix,
			JSSDKCDNURL:          SSOSetting.JSSDKCDNURL,
			StateJWTSecret:       SSOSetting.StateJWTSecret,
			AutoLinkProviderKeys: SSOSetting.AutoLinkProviderKeys,
			AllowedCallbackURLs:  SSOSetting.AllowedCallbackURLs,
		}
		config := sso.Config{
			Name:         SSOConf.Name,
			ClientID:     SSOConf.ClientID,
			ClientSecret: SSOConf.ClientSecret,
			Scope:        strings.Split(SSOConf.Scope, ","),
		}
		return sso.NewProvider(setting, config)
	case "VerifyCodeStore":
		tConfig := config.GetTenantConfig(r)
		return verifycode.NewStore(
			db.NewSQLBuilder("auth", tConfig.AppName),
			db.NewSQLExecutor(r.Context(), db.NewContextWithContext(r.Context(), tConfig)),
			logging.CreateLogger(r, "verify_code", createLoggerMaskFormatter(r)),
		)
	case "UserVerifyCodeSenderFactory":
		tConfig := config.GetTenantConfig(r)
		return NewDefaultUserVerifyCodeSenderFactory(tConfig)
	default:
		return nil
	}
}

func createLoggerMaskFormatter(r *http.Request) logrus.Formatter {
	tConfig := config.GetTenantConfig(r)
	return logging.CreateMaskFormatter(tConfig.DefaultSensitiveLoggerValues(), &logrus.TextFormatter{})
}

type UserVerifyCodeSenderFactory interface {
	NewCodeSender(key string) userverify.CodeSender
}

type DefaultUserVerifyCodeSenderFactory struct {
	CodeSenderMap map[string]userverify.CodeSender
}

func NewDefaultUserVerifyCodeSenderFactory(c config.TenantConfiguration) UserVerifyCodeSenderFactory {
	userVerifyConfig := c.UserVerify
	f := DefaultUserVerifyCodeSenderFactory{
		CodeSenderMap: map[string]userverify.CodeSender{},
	}
	for _, keyConfig := range userVerifyConfig.KeyConfigs {
		var codeSender userverify.CodeSender
		switch keyConfig.Provider {
		case "smtp":
			codeSender = &userverify.EmailCodeSender{
				AppName:       c.AppName,
				Key:           keyConfig.Key,
				Config:        userVerifyConfig,
				Dialer:        mail.NewDialer(c.SMTP),
				CodeGenerator: userverify.NewCodeGenerator(keyConfig.CodeFormat),
			}
		case "twilio":
			codeSender = &userverify.TwilioCodeSender{
				AppName:       c.AppName,
				Key:           keyConfig.Key,
				Config:        userVerifyConfig,
				TwilioClient:  sms.NewTwilioClient(c.Twilio),
				CodeGenerator: userverify.NewCodeGenerator(keyConfig.CodeFormat),
			}
		case "nexmo":
			codeSender = &userverify.NexmoCodeSender{
				AppName:       c.AppName,
				Key:           keyConfig.Key,
				Config:        userVerifyConfig,
				NexmoClient:   sms.NewNexmoClient(c.Nexmo),
				CodeGenerator: userverify.NewCodeGenerator(keyConfig.CodeFormat),
			}
		default:
			panic(errors.New("invalid user verify provider: " + keyConfig.Provider))
		}
		f.CodeSenderMap[keyConfig.Key] = codeSender
	}

	return &f
}

func (d *DefaultUserVerifyCodeSenderFactory) NewCodeSender(key string) userverify.CodeSender {
	return d.CodeSenderMap[key]
}
