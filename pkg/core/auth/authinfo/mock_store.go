package authinfo

import (
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

// MockStore is the memory implementation of authinfo store
type MockStore struct {
	AuthInfoMap map[string]AuthInfo
}

// NewMockStore create mock store with empty AuthInfoMap
func NewMockStore() *MockStore {
	return NewMockStoreWithAuthInfoMap(map[string]AuthInfo{})
}

// NewMockStoreWithAuthInfoMap create mock store with AuthInfoMap fixture
func NewMockStoreWithAuthInfoMap(authInfoMap map[string]AuthInfo) *MockStore {
	return &MockStore{
		AuthInfoMap: authInfoMap,
	}
}

// CreateAuth creates AuthInfo in AuthInfoMap.
func (s *MockStore) CreateAuth(authinfo *AuthInfo) error {
	if _, existed := s.AuthInfoMap[authinfo.ID]; existed {
		return skydb.ErrUserDuplicated
	}
	s.AuthInfoMap[authinfo.ID] = *authinfo
	return nil
}

// GetAuth get AuthInfo in AuthInfoMap.
func (s *MockStore) GetAuth(id string, authinfo *AuthInfo) error {
	u, existed := s.AuthInfoMap[id]
	if !existed {
		return skydb.ErrUserNotFound
	}

	*authinfo = u
	return nil
}

// UpdateAuth update AuthInfo in AuthInfoMap.
func (s *MockStore) UpdateAuth(authinfo *AuthInfo) error {
	if _, ok := s.AuthInfoMap[authinfo.ID]; !ok {
		return skydb.ErrUserNotFound
	}

	s.AuthInfoMap[authinfo.ID] = *authinfo
	return nil
}

// DeleteAuth delete AuthInfo in AuthInfoMap.
func (s *MockStore) DeleteAuth(id string) error {
	if _, ok := s.AuthInfoMap[id]; !ok {
		return skydb.ErrUserNotFound
	}
	delete(s.AuthInfoMap, id)
	return nil
}

// AssignRoles updates roles of authinfo in AuthInfoMap.
func (s *MockStore) AssignRoles(userIDs []string, roles []string) error {
	for _, userID := range userIDs {
		authInfo, existed := s.AuthInfoMap[userID]
		if !existed {
			continue
		}
		roleSet := make(map[string]interface{})
		for _, role := range authInfo.Roles {
			roleSet[role] = struct{}{}
		}
		for _, role := range roles {
			if _, ok := roleSet[role]; !ok {
				roleSet[role] = struct{}{}
				authInfo.Roles = append(authInfo.Roles, role)
				s.AuthInfoMap[userID] = authInfo
			}
		}
	}
	return nil
}

// GetRoles accepts array of userID, and return corresponding roles from the
// AuthInfoMap
func (s *MockStore) GetRoles(userIDs []string) (roleMap map[string][]string, err error) {
	roleMap = map[string][]string{}
	for _, userID := range userIDs {
		authInfo, existed := s.AuthInfoMap[userID]
		if existed {
			roleMap[userID] = authInfo.Roles
		}
	}
	return roleMap, nil
}

// RevokeRoles accepts array of roles and userID, the supplied roles in
// AuthInfoMap will be revoked from all passed in users
func (s *MockStore) RevokeRoles(userIDs []string, roles []string) error {
	for _, userID := range userIDs {
		authInfo, existed := s.AuthInfoMap[userID]
		if !existed {
			continue
		}
		role2revoke := make(map[string]interface{})
		for _, role := range roles {
			role2revoke[role] = struct{}{}
		}
		oldRoles := append(authInfo.Roles[:0:0], authInfo.Roles...)
		authInfo.Roles = []string{}
		for _, r := range oldRoles {
			if _, existed := role2revoke[r]; !existed {
				authInfo.Roles = append(authInfo.Roles, r)
			}
		}
		s.AuthInfoMap[userID] = authInfo
	}
	return nil
}
