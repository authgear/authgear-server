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

func (s *AuthzService) AddAuthorizedUser(appID string, userID string) error {
	// FIXME(authz): add authorized user to app
	return nil
}
