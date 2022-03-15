package deps

import (
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
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
		appCtx, err := m.ConfigSource.ProvideContext(r)
		if err != nil {
			if errors.Is(err, configsource.ErrAppNotFound) {
				http.Error(w, configsource.ErrAppNotFound.Error(), http.StatusNotFound)
			} else {
				logger.WithError(err).Error("failed to resolve config")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		ap := m.RootProvider.NewAppProvider(r.Context(), appCtx)
		r = r.WithContext(withProvider(r.Context(), ap))
		next.ServeHTTP(w, r)
	})
}
