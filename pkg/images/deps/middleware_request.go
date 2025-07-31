package deps

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
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
		logger.Debug(ctx, "serving request",
			slog.String("request.host", r.Host),
			slog.String("request.url", r.URL.String()),
			slog.String("request.header.host", r.Header.Get("Host")),
			slog.String("request.header.x-forwarded-host", r.Header.Get("X-Forwarded-Host")),
			slog.String("request.header.x-original-host", r.Header.Get("X-Original-Host")),
		)
		err := m.ConfigSource.ProvideContext(ctx, r, func(ctx context.Context, appCtx *config.AppContext) error {
			ctx, ap := m.RootProvider.NewAppProvider(ctx, appCtx)
			ctx = withProvider(ctx, ap)
			otelauthgear.SetProjectID(ctx, string(appCtx.Config.AppConfig.ID))

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
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
