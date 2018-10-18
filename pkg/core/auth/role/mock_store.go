package role

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// MockStore is the memory implementation of role store
type MockStore struct {
	RoleMap map[string]Role
}

// NewMockStore creates a new mock instance
func NewMockStore() *MockStore {
	return NewMockStoreWithRoleMap(map[string]Role{})
}

// NewMockStoreWithRoleMap create mock store with RoleMap fixture
func NewMockStoreWithRoleMap(roleMap map[string]Role) *MockStore {
	return &MockStore{
		RoleMap: roleMap,
	}
}

// CreateRoles create Role in RoleMap
func (m *MockStore) CreateRoles(roles []string) error {
	for _, role := range roles {
		if _, existed := m.RoleMap[role]; existed {
			return skyerr.NewError(skyerr.Duplicated,
				fmt.Sprintf("Duplicated roles %v", role))
		}
		m.RoleMap[role] = Role{
			Name: role,
		}
	}

	return nil
}

// QueryRoles query roles from RoleMap by name
func (m *MockStore) QueryRoles(roles []string) ([]Role, error) {
	if roles == nil {
		return nil, nil
	}

	if len(roles) == 0 {
		return []Role{}, nil
	}

	existedRoles := []Role{}
	for _, name := range roles {
		role, existed := m.RoleMap[name]
		if existed {
			existedRoles = append(existedRoles, role)
		}
	}
	return existedRoles, nil
}

// GetDefaultRoles returns default roles from map
func (m *MockStore) GetDefaultRoles() ([]string, error) {
	defaultRoles := []string{}
	for _, role := range m.RoleMap {
		if role.IsDefault {
			defaultRoles = append(defaultRoles, role.Name)
		}
	}
	return defaultRoles, nil
}

// SetAdminRoles set role type to true
func (m *MockStore) SetAdminRoles(roles []string) error {
	return m.setRoleType(roles, "is_admin")
}

func (m *MockStore) setRoleType(roles []string, col string) error {
	isAdmin := false
	isDefault := false

	if col == "is_admin" {
		isAdmin = true
	} else if col == "is_default" {
		isDefault = true
	}

	// reset current roles
	for _, role := range m.RoleMap {
		role.IsAdmin = false
		role.IsDefault = false
	}

	// update roles
	for _, role := range roles {
		m.RoleMap[role] = Role{
			Name:      role,
			IsAdmin:   isAdmin,
			IsDefault: isDefault,
		}
	}

	return nil
}
