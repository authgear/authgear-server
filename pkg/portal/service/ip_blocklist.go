package service

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/geoip"
)

type IPBlocklistService struct {
}

func NewIPBlocklistService() *IPBlocklistService {
	return &IPBlocklistService{}
}

func (s *IPBlocklistService) CheckIP(ctx context.Context, ipAddress string, cidrs []string, countryCodes []string) bool {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		// CIDRs should be validated before passing into this method
		panic(fmt.Errorf("invalid IP address: %s", ipAddress))
	}

	if s.isIPBlocked(ip, cidrs) {
		return true
	}

	if s.isCountryBlocked(ipAddress, countryCodes) {
		return true
	}

	return false
}

func (s *IPBlocklistService) isIPBlocked(ip net.IP, cidrs []string) bool {
	for _, cidrStr := range cidrs {
		_, cidrNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			// CIDRs should be validated before passing into this method
			panic(err)
		}

		if cidrNet.Contains(ip) {
			return true
		}
	}
	return false
}

func (s *IPBlocklistService) isCountryBlocked(remoteIP string, countryCodes []string) bool {
	if len(countryCodes) == 0 {
		return false
	}
	if info, ok := geoip.IPString(remoteIP); ok {
		countryCode := info.CountryCode
		for _, blocked := range countryCodes {
			if strings.EqualFold(countryCode, blocked) {
				return true
			}
		}
	}
	return false
}
