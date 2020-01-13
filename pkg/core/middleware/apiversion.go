package middleware

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/apiversion"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var ErrIncompatibleAPIVersion = skyerr.NotFound.WithReason("IncompatibleAPIVersion").New("incompatible API version")

// APIVersionMiddleware compares the API version with own version.
type APIVersionMiddleware struct {
	// APIVersionName tells how to extract the API version from mux.Vars.
	APIVersionName string
	// MajorVersion determines the supported major version.
	// It should be apiversion.MajorVersion in normal use case.
	MajorVersion int
	// MinorVersion determines the supported minor version.
	// It should be apiversion.MinorVersion in normal use case.
	MinorVersion int
}

func (m *APIVersionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiVersion := mux.Vars(r)[m.APIVersionName]
		major, minor, ok := apiversion.Parse(apiVersion)

		if !ok || major != m.MajorVersion || minor > m.MinorVersion {
			handler.WriteResponse(w, handler.APIResponse{
				Error: ErrIncompatibleAPIVersion,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
