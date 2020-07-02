package deps

import (
	"github.com/authgear/authgear-server/pkg/core/errors"
	"net/http"

	configsource "github.com/authgear/authgear-server/pkg/auth/config/source"
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
			if errors.Is(err, configsource.ErrAppNotFound) {
				http.Error(w, configsource.ErrAppNotFound.Error(), http.StatusNotFound)
			} else {
				logger.WithError(err).Error("failed to resolve config")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		rp := m.RootProvider.NewRequestProvider(r, config)
		r = r.WithContext(withProvider(r.Context(), rp))
		next.ServeHTTP(w, r)
	})
}
