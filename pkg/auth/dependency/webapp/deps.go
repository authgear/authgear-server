package webapp

import (
	"github.com/google/wire"
	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/deps"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

func ProvideValidateProvider(tConfig *config.TenantConfiguration) ValidateProvider {
	return &ValidateProviderImpl{
		Validator:                       validator,
		LoginIDConfiguration:            tConfig.AppConfig.Identity.LoginID,
		CountryCallingCodeConfiguration: tConfig.AppConfig.AuthUI.CountryCallingCode,
	}
}

func ProvideRenderProvider(
	saup deps.StaticAssetURLPrefix,
	config *config.TenantConfiguration,
	templateEngine *template.Engine,
	passwordChecker *audit.PasswordChecker,
) RenderProvider {
	return &RenderProviderImpl{
		StaticAssetURLPrefix:        string(saup),
		IdentityConfiguration:       config.AppConfig.Identity,
		AuthenticationConfiguration: config.AppConfig.Authentication,
		AuthUIConfiguration:         config.AppConfig.AuthUI,
		PasswordChecker:             passwordChecker,
		TemplateEngine:              templateEngine,
	}
}

var DependencySet = wire.NewSet(
	ProvideValidateProvider,
	ProvideRenderProvider,
	wire.Struct(new(StateStoreImpl), "*"),
	wire.Bind(new(StateStore), new(*StateStoreImpl)),
)

func ProvideCSPMiddleware(tConfig *config.TenantConfiguration) mux.MiddlewareFunc {
	m := &CSPMiddleware{Clients: tConfig.AppConfig.Clients}
	return m.Handle
}

func ProvideStateMiddleware(stateStore StateStore) mux.MiddlewareFunc {
	m := &StateMiddleware{StateStore: stateStore}
	return m.Handle
}

func ProvideClientIDMiddleware(tConfig *config.TenantConfiguration) mux.MiddlewareFunc {
	m := &ClientIDMiddleware{TenantConfig: tConfig}
	return m.Handle
}
