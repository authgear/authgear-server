package service

import (
	"context"
	"errors"
	"net"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

//go:generate mockgen -source=default_domain.go -destination=default_domain_mock_test.go -package service

var ErrHostSuffixNotConfigured = errors.New("host suffix not configured")

type DefaultDomainDomainService interface {
	CreateDomain(ctx context.Context, appID string, domain string, isVerified bool, isCustom bool) (*apimodel.Domain, error)
}

type DefaultDomainService struct {
	AppHostSuffixes config.AppHostSuffixes
	AppConfig       *portalconfig.AppConfig
	Domains         DefaultDomainDomainService
}

// GetLatestAppHost does not need connection.
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

// CreateAllDefaultDomains assume acquired connection.
func (s *DefaultDomainService) CreateAllDefaultDomains(ctx context.Context, appID string) error {
	if s.AppConfig.HostSuffix == "" {
		return ErrHostSuffixNotConfigured
	}

	suffixes := []string{s.AppConfig.HostSuffix}
	for _, hostSuffix := range s.AppHostSuffixes {
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
		_, err := s.Domains.CreateDomain(ctx, appID, domain, true, false)
		if err != nil {
			return err
		}
	}

	return nil
}
