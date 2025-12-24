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

		remoteIPStr := string(m.RemoteIP)
		remoteIP := net.ParseIP(remoteIPStr)
		if remoteIP == nil {
			next.ServeHTTP(w, r)
			return
		}

		defaultAction := m.Config.IPFilter.DefaultAction

		action := m.evaluate(remoteIP, remoteIPStr, defaultAction)
		if action == config.IPFilterActionDeny {
			http.Error(w, "Your IP is not allowed to access this resource", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *IPFilterMiddleware) evaluate(remoteIP net.IP, remoteIPStr string, defaultAction config.IPFilterAction) config.IPFilterAction {
	for _, rule := range m.Config.IPFilter.Rules {
		if m.isRuleMatched(remoteIP, remoteIPStr, rule) {
			return rule.Action
		}
	}

	return defaultAction
}

func (m *IPFilterMiddleware) isRuleMatched(remoteIP net.IP, remoteIPStr string, rule *config.IPFilterRule) bool {
	// An empty source means the rule will never be matched
	if len(rule.Source.CIDRs) == 0 && len(rule.Source.GeoLocationCodes) == 0 {
		return false
	}

	cidrMatch := false
	if len(rule.Source.CIDRs) > 0 {
		cidrMatch = matchCIDRs(remoteIP, rule.Source.CIDRs)
	}

	geoMatch := false
	if len(rule.Source.GeoLocationCodes) > 0 {
		geoMatch = m.isGeoIPBlocked(remoteIPStr, rule.Source.GeoLocationCodes)
	}

	return cidrMatch || geoMatch
}

func (m *IPFilterMiddleware) isGeoIPBlocked(remoteIP string, countryCodes []string) bool {
	if len(countryCodes) == 0 {
		return false
	}

	if info, ok := geoip.IPString(remoteIP); ok {
		return matchCountryCode(info.CountryCode, countryCodes)
	}

	return false
}

func matchCIDRs(ip net.IP, cidrs []string) bool {
	for _, cidrStr := range cidrs {
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

func matchCountryCode(countryCode string, countryCodes []string) bool {
	for _, blocked := range countryCodes {
		if strings.EqualFold(countryCode, blocked) {
			return true
		}
	}
	return false
}
