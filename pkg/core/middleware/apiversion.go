package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var IncompatibleAPIVersion = skyerr.NotFound.WithReason("IncompatibleAPIVersion")

// APIVersionMiddleware compares the API version with own version.
type APIVersionMiddleware struct {
	// APIVersionName tells how to extract the API version from mux.Vars.
	APIVersionName string
	// SupportedVersions determines the full list of supported versions.
	SupportedVersions []string
	// SupportedVersionsJSON is the JSON encoding of SupportedVersions.
	SupportedVersionsJSON string
}

func (m *APIVersionMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiVersion := mux.Vars(r)[m.APIVersionName]

		for _, supported := range m.SupportedVersions {
			if apiVersion == supported {
				next.ServeHTTP(w, r)
				return
			}
		}

		handler.WriteResponse(w, handler.APIResponse{
			Error: IncompatibleAPIVersion.New(
				fmt.Sprintf("expected API versions: %s", m.SupportedVersionsJSON),
			),
		})
	})
}
