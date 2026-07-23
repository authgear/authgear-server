package deps

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var RequestMiddlewareLogger = slogutil.NewLogger("request")

type RequestMiddleware struct {
	RootProvider *RootProvider
	ConfigSource *configsource.ConfigSource
}

func (m *RequestMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := RequestMiddlewareLogger.GetLogger(ctx)
		err := m.ConfigSource.ProvideContext(ctx, r, func(ctx context.Context, appCtx *config.AppContext) error {
			ctx, ap := m.RootProvider.NewAppProvider(ctx, appCtx)
			ctx = WithAppProvider(ctx, ap)
			// We create a new labeler from the global labeler.
			// Add project specific labels to it.
			// So metrics produced under this middleware have project specific labels.
			// And metrics produced before this middleware have no project specific labels.
			ctx = otelutil.ContextWithClonedLabeler(ctx)
			otelauthgear.SetProjectID(ctx, string(appCtx.Config.AppConfig.ID))

			otelauthgear.ServeHTTPWithRequestCountMetric(ctx, w, r, next)
			return nil
		})
		if err != nil {
			if errors.Is(err, configsource.ErrAppNotFound) {
				http.Error(w, configsource.ErrAppNotFound.Error(), http.StatusNotFound)
			} else {
				logger.WithError(err).Error(ctx, "failed to resolve config")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
	})
}
