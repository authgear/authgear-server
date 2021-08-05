//+build wireinject

package smtp

import (
	"context"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func NewTranslationService(ctx context.Context, app *model.App) *translation.Service {
	panic(wire.Build(
		ProvideStaticAssetResolver,
		ProvideResourceManager,
		ProvideDefaultLanguageTag,
		ProvideSupportedLanguageTags,

		translation.DependencySet,
		template.DependencySet,

		wire.Bind(new(template.ResourceManager), new(*resource.Manager)),
		wire.Bind(new(translation.StaticAssetResolver), new(*NoopStaticAssetResolver)),
	))
}
