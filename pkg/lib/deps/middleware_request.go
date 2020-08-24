package deps

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/errorutil"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

type RequestMiddleware struct {
	RootProvider *RootProvider
	ConfigSource *configsource.ConfigSource
}

func (m *RequestMiddleware) Handle(next http.Handler) http.Handler {
	logger := m.RootProvider.LoggerFactory.New("request")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appCtx, err := m.ConfigSource.ProvideContext(r)
		if err != nil {
			if errorutil.Is(err, configsource.ErrAppNotFound) {
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
