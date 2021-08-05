package smtp

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
)

type NoopStaticAssetResolver struct{}

func (r *NoopStaticAssetResolver) StaticAssetURL(id string) (url string, err error) {
	panic("NoopStaticAssetResolver is not supposed to be reachable")
}

func ProvideStaticAssetResolver() *NoopStaticAssetResolver {
	return &NoopStaticAssetResolver{}
}

func ProvideResourceManager(app *model.App) *resource.Manager {
	return app.Context.Resources
}

func ProvideDefaultLanguageTag(app *model.App) template.DefaultLanguageTag {
	return template.DefaultLanguageTag(*app.Context.Config.AppConfig.Localization.FallbackLanguage)
}

func ProvideSupportedLanguageTags(app *model.App) template.SupportedLanguageTags {
	return template.SupportedLanguageTags(app.Context.Config.AppConfig.Localization.SupportedLanguages)
}
