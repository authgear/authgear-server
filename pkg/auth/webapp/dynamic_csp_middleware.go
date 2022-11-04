package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type DynamicCSPMiddleware struct {
	HTTPConfig        *config.HTTPConfig
	WebAppCDNHost     config.WebAppCDNHost
	AllowInlineScript bool
}

func (m *DynamicCSPMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := web.GetCSPNonce(r.Context())
		cspDirectives, err := web.CSPDirectives(m.HTTPConfig.PublicOrigin, nonce, string(m.WebAppCDNHost), m.AllowInlineScript)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Security-Policy", httputil.CSPJoin(cspDirectives))
		next.ServeHTTP(w, r)
	})
}
