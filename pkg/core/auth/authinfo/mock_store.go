package authinfo

import (
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

// MockStore is the memory implementation of authinfo store
type MockStore struct {
	AuthInfoMap map[string]AuthInfo
}

// NewMockStore create mock store with empty AuthInfoMap
func NewMockStore() *MockStore {
	return NewMockStoreWithAuthInfoMap(map[string]AuthInfo{})
}

// NewMockStoreWithUser create mock store with user
func NewMockStoreWithUser(userID string) *MockStore {
	return NewMockStoreWithAuthInfoMap(map[string]AuthInfo{
		userID: AuthInfo{
			ID: userID,
		},
	})
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
		return errors.New("dupliated auth info")
	}
	s.AuthInfoMap[authinfo.ID] = *authinfo
	return nil
}

// GetAuth get AuthInfo in AuthInfoMap.
func (s *MockStore) GetAuth(id string, authinfo *AuthInfo) error {
	u, existed := s.AuthInfoMap[id]
	if !existed {
		return ErrNotFound
	}

	*authinfo = u
	return nil
}

// UpdateAuth update AuthInfo in AuthInfoMap.
func (s *MockStore) UpdateAuth(authinfo *AuthInfo) error {
	if _, ok := s.AuthInfoMap[authinfo.ID]; !ok {
		return ErrNotFound
	}

	s.AuthInfoMap[authinfo.ID] = *authinfo
	return nil
}

// DeleteAuth delete AuthInfo in AuthInfoMap.
func (s *MockStore) DeleteAuth(id string) error {
	if _, ok := s.AuthInfoMap[id]; !ok {
		return ErrNotFound
	}
	delete(s.AuthInfoMap, id)
	return nil
}
