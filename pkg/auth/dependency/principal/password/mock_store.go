package password

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
)

type MockStore struct {
	PrincipalMap map[string]Principal
}

func NewMockStore() Store {
	return &MockStore{
		PrincipalMap: map[string]Principal{},
	}
}

func (s *MockStore) CreatePrincipal(principal Principal) error {
	if _, ok := s.PrincipalMap[principal.ID]; ok {
		return ErrLoginIDAlreadyUsed
	}

	for _, p := range s.PrincipalMap {
		if p.Realm != principal.Realm {
			continue
		}
		if p.LoginID == principal.LoginID || p.UniqueKey == principal.UniqueKey {
			return ErrLoginIDAlreadyUsed
		}
	}

	s.PrincipalMap[principal.ID] = principal

	return nil
}

func (s *MockStore) GetPrincipals(loginIDKey string, loginID string, realm *string) ([]*Principal, error) {
	principals := []*Principal{}
	for _, p := range s.PrincipalMap {
		if p.LoginIDKey == loginIDKey && p.LoginID == loginID {
			if realm != nil && p.Realm != *realm {
				continue
			}
			found := p
			principals = append(principals, &found)
		}
	}

	return principals, nil
}

func (s *MockStore) GetPrincipalByID(principalID string) (principal.Principal, error) {
	if p, ok := s.PrincipalMap[principalID]; ok {
		return &p, nil
	}

	return nil, principal.ErrNotFound
}

func (s *MockStore) GetPrincipalsByUserID(userID string) ([]*Principal, error) {
	principals := []*Principal{}
	for _, p := range s.PrincipalMap {
		if p.UserID == userID {
			found := p
			principals = append(principals, &found)
		}
	}

	return principals, nil
}

func (s *MockStore) GetPrincipalsByClaim(claimName string, claimValue string) ([]*Principal, error) {
	principals := []*Principal{}
	for _, p := range s.PrincipalMap {
		if v, ok := p.ClaimsValue[claimName]; ok {
			if v == claimValue {
				found := p
				principals = append(principals, &found)
			}
		}
	}

	return principals, nil

}

func (s *MockStore) UpdatePassword(principal *Principal, password string) (err error) {
	return nil
}

var (
	_ Store = &MockStore{}
)
