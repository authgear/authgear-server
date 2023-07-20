package service

import (
	"errors"
	"net"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

//go:generate mockgen -source=default_domain.go -destination=default_domain_mock_test.go -package service

var ErrHostSuffixNotConfigured = errors.New("host suffix not configured")

type DefaultDomainDomainService interface {
	CreateDomain(appID string, domain string, isVerified bool, isCustom bool) (*model.Domain, error)
}

type DefaultDomainService struct {
	AppConfig *portalconfig.AppConfig
	Domains   DefaultDomainDomainService
}

func (s *DefaultDomainService) GetLatestAppHost(appID string) (string, error) {
	if s.AppConfig.HostSuffix == "" {
		return "", ErrHostSuffixNotConfigured
	}
	return s.makeHost(appID, s.AppConfig.HostSuffix), nil
}

func (s *DefaultDomainService) makeHost(appID string, suffix string) string {
	return appID + suffix
}

func (s *DefaultDomainService) hostToDomain(host string) string {
	h, _, err := net.SplitHostPort(host)
	if err == nil {
		return h
	}
	return host
}

func (s *DefaultDomainService) CreateAllDefaultDomains(appID string) error {
	if s.AppConfig.HostSuffix == "" {
		return ErrHostSuffixNotConfigured
	}

	suffixes := []string{s.AppConfig.HostSuffix}
	for _, hostSuffix := range s.AppConfig.HostSuffixes {
		suffixes = append(suffixes, hostSuffix)
	}

	suffixes = slice.Deduplicate(suffixes)

	hosts := make([]string, len(suffixes))
	for i, suffix := range suffixes {
		host := s.makeHost(appID, suffix)
		hosts[i] = host
	}

	domains := make([]string, len(hosts))
	for i, host := range hosts {
		domain := s.hostToDomain(host)
		domains[i] = domain
	}

	for _, domain := range domains {
		_, err := s.Domains.CreateDomain(appID, domain, true, false)
		if err != nil {
			return err
		}
	}

	return nil
}
