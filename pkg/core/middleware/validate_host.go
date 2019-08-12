package middleware

import (
	"net/http"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/utils"

	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

// ValidateHostMiddleware validate incoming request has correct Host header
type ValidateHostMiddleware struct {
	ValidHosts string
}

func (m ValidateHostMiddleware) Handle(next http.Handler) http.Handler {
	validateHosts := strings.Split(m.ValidHosts, ",")
	for i, host := range validateHosts {
		validateHosts[i] = strings.TrimSpace(host)
	}

	if len(validateHosts) == 1 && validateHosts[0] == "" {
		// skip validation if no host is provided
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := coreHttp.GetHost(r)
		if !utils.StringSliceContains(validateHosts, host) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte{})
		}

		next.ServeHTTP(w, r)
	})
}
