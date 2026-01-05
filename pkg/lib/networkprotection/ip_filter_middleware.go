package networkprotection

import (
	"net"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IPFilterMiddleware struct {
	RemoteIP httputil.RemoteIP
	Config   *config.NetworkProtectionConfig
}

func (m *IPFilterMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Config == nil || m.Config.IPFilter == nil {
			next.ServeHTTP(w, r)
			return
		}
		remoteIP := net.ParseIP(string(m.RemoteIP))
		action := Evaluate(m.Config, remoteIP)
		if action == config.IPFilterActionDeny {
			http.Error(w, "Your IP is not allowed to access this resource", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
