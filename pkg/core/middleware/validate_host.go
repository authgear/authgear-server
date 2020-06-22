package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/utils"
	"github.com/skygeario/skygear-server/pkg/httputil"
)

// FIXME: remove this
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
		host := httputil.GetHost(r, true)
		if !utils.StringSliceContains(validateHosts, host) {
			http.Error(w, fmt.Sprintf("invalid host: %s", host), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
