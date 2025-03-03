package deps

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type RequestMiddleware struct {
	RootProvider *RootProvider
	ConfigSource *configsource.ConfigSource
}

func (m *RequestMiddleware) Handle(next http.Handler) http.Handler {
	logger := m.RootProvider.LoggerFactory.New("request")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(map[string]interface{}{
			"request.host":                    r.Host,
			"request.url":                     r.URL.String(),
			"request.header.host":             r.Header.Get("Host"),
			"request.header.x-forwarded-host": r.Header.Get("X-Forwarded-Host"),
			"request.header.x-original-host":  r.Header.Get("X-Original-Host"),
		}).Debug("serving request")
		err := m.ConfigSource.ProvideContext(r.Context(), r, func(ctx context.Context, appCtx *config.AppContext) error {
			r = r.WithContext(ctx)

			ap := m.RootProvider.NewAppProvider(r.Context(), appCtx)
			r = r.WithContext(WithAppProvider(r.Context(), ap))

			otelauthgear.SetProjectID(r.Context(), string(appCtx.Config.AppConfig.ID))
			next.ServeHTTP(w, r)
			return nil
		})
		if err != nil {
			if errors.Is(err, configsource.ErrAppNotFound) {
				http.Error(w, configsource.ErrAppNotFound.Error(), http.StatusNotFound)
			} else {
				err = errorutil.ForceLogging(err)

				otelauthgear.TrackContextCanceled(r.Context(), err, r, bool(m.RootProvider.EnvironmentConfig.TrustProxy))

				// Our logging mechanism is not context-aware.
				// We explicitly attach context here because it was the position we observed the log.
				logger.WithContext(r.Context()).WithError(err).Error("failed to resolve config")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}
	})
}
