package service

type AuthzConfigService interface {
	ListAllAppIDs() ([]string, error)
}

type AuthzService struct {
	AppConfigs AuthzConfigService
}

func (s *AuthzService) ListAuthorizedApps(userID string) ([]string, error) {
	// FIXME(authz): extract authorized app from user labels
	appIDs, err := s.AppConfigs.ListAllAppIDs()
	return appIDs, err
}
