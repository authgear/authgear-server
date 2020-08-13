package deps

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/errorutil"

	configsource "github.com/authgear/authgear-server/pkg/lib/config/source"
)

type RequestMiddleware struct {
	RootProvider *RootProvider
	ConfigSource configsource.Source
}

func (m *RequestMiddleware) Handle(next http.Handler) http.Handler {
	logger := m.RootProvider.LoggerFactory.New("request")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config, err := m.ConfigSource.ProvideConfig(r.Context(), r)
		if err != nil {
			if errorutil.Is(err, configsource.ErrAppNotFound) {
				http.Error(w, configsource.ErrAppNotFound.Error(), http.StatusNotFound)
			} else {
				logger.WithError(err).Error("failed to resolve config")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		ap := m.RootProvider.NewAppProvider(r.Context(), config)
		r = r.WithContext(withProvider(r.Context(), ap))
		next.ServeHTTP(w, r)
	})
}
