package middleware

import (
	"encoding/json"
	"net/http"

	nextSkyerr "github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// RecoveredResponse is interface for RecoverMiddleware to write response
type RecoveredResponse struct {
	Err skyerr.Error `json:"error,omitempty"`
}

// RecoverMiddleware recover from panic
type RecoverMiddleware struct{}

func (m RecoverMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				err := nextSkyerr.ErrorFromRecoveringPanic(r)
				httpStatus := nextSkyerr.ErrorDefaultStatusCode(err)

				// TODO: log

				response := RecoveredResponse{Err: err}
				encoder := json.NewEncoder(w)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(httpStatus)
				encoder.Encode(response)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
