package service

import "github.com/authgear/authgear-server/pkg/lib/config/configsource"

type AuthzService struct {
	ConfigSource *configsource.ConfigSource
}

func (s *AuthzService) ListAuthorizedApps(userID string) ([]string, error) {
	// FIXME(authz): extract authorized app from user labels
	appIDs, err := s.ConfigSource.AppIDResolver.AllAppIDs()
	return appIDs, err
}
