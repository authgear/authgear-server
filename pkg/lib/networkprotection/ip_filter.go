package networkprotection

import (
	"fmt"
	"net"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/geoip"
)

func Evaluate(cfg *config.NetworkProtectionConfig, remoteIP net.IP) config.IPFilterAction {
	if cfg == nil || cfg.IPFilter == nil {
		panic("networkprotection: unexpected nil config")
	}

	defaultAction := cfg.IPFilter.DefaultAction

	if remoteIP == nil {
		return defaultAction
	}
	remoteIPStr := remoteIP.String()

	for _, rule := range cfg.IPFilter.Rules {
		if isRuleMatched(remoteIP, remoteIPStr, rule) {
			return rule.Action
		}
	}

	return defaultAction
}

func isRuleMatched(remoteIP net.IP, remoteIPStr string, rule *config.IPFilterRule) bool {
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
		geoMatch = isGeoIPBlocked(remoteIPStr, rule.Source.GeoLocationCodes)
	}

	return cidrMatch || geoMatch
}

func isGeoIPBlocked(remoteIP string, countryCodes []string) bool {
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
