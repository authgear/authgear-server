package oauth

type AuthorizationService struct {
	Store AuthorizationStore
}

func (s *AuthorizationService) GetByID(id string) (*Authorization, error) {
	return s.Store.GetByID(id)
}

func (s *AuthorizationService) ListByUser(userID string, filters ...Filter) ([]*Authorization, error) {
	as, err := s.Store.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	filtered := []*Authorization{}
	for _, a := range as {
		keep := true
		for _, f := range filters {
			if !f.Keep(a) {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, a)
		}
	}

	return filtered, nil
}

func (s *AuthorizationService) Delete(a *Authorization) error {
	return s.Store.Delete(a)
}
