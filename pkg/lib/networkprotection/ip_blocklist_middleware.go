package networkprotection

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type IPBlocklistMiddleware struct {
	RemoteIP httputil.RemoteIP
	Config   *config.NetworkProtectionConfig
}

func (m *IPBlocklistMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Config != nil && m.Config.IPBlocklist != nil {
			remoteIP := string(m.RemoteIP)
			ip := net.ParseIP(remoteIP)

			if ip != nil {
				if m.isIPBlocked(ip) || m.isCountryBlocked(remoteIP) {
					http.Error(w, "Your IP is not allowed to access this resource", http.StatusForbidden)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *IPBlocklistMiddleware) isIPBlocked(ip net.IP) bool {
	for _, cidrStr := range m.Config.IPBlocklist.CIDRs {
		_, cidrNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			panic(fmt.Errorf("failed to parse cidr: %w", err))
		}

		if cidrNet.Contains(ip) {
			return true
		}
	}
	return false
}

func (m *IPBlocklistMiddleware) isCountryBlocked(remoteIP string) bool {
	if len(m.Config.IPBlocklist.CountryCodes) == 0 {
		return false
	}
	if info, ok := geoip.IPString(remoteIP); ok {
		countryCode := info.CountryCode
		for _, blocked := range m.Config.IPBlocklist.CountryCodes {
			if strings.EqualFold(countryCode, blocked) {
				return true
			}
		}
	}
	return false
}
