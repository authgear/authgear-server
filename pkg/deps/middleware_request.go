package deps

import (
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"net/http"

	configsource "github.com/skygeario/skygear-server/pkg/auth/config/source"
)

type RequestMiddleware struct {
	RootContainer *RootContainer
	ConfigSource  configsource.Source
}

func (m *RequestMiddleware) Handle(next http.Handler) http.Handler {
	logger := m.RootContainer.LoggerFactory.New("request")

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

		requestContainer := m.RootContainer.NewRequestContainer(r.Context(), r, config)
		r = r.WithContext(WithRequestContainer(r.Context(), requestContainer))
		next.ServeHTTP(w, r)
	})
}
