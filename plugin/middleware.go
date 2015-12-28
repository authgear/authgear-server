package plugin

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/skyerr"
)

func AvailabilityMiddleware(next http.Handler, initContext *InitContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !initContext.IsReady() {
			log.Errorf("Request cannot be handled because plugins are unavailable at the moment.")
			err := skyerr.NewError(skyerr.PluginUnavailable, "plugins are unavailable at the moment")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(struct {
				Error skyerr.Error `json:"error"`
			}{
				Error: err,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
