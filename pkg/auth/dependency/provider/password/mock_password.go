package password

import (
	"reflect"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"golang.org/x/crypto/bcrypt"
)

// MockProvider is the memory implementation of password provider
type MockProvider struct {
	Provider
	PrincipalMap map[string]Principal
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider() *MockProvider {
	return NewMockProviderWithPrincipalMap(map[string]Principal{})
}

// NewMockProviderWithPrincipalMap creates a new instance of mock provider with PrincipalMap
func NewMockProviderWithPrincipalMap(principalMap map[string]Principal) *MockProvider {
	return &MockProvider{
		PrincipalMap: principalMap,
	}
}

// CreatePrincipal creates principal in PrincipalMap
func (m *MockProvider) CreatePrincipal(principal Principal) error {
	if _, existed := m.PrincipalMap[principal.ID]; existed {
		return skydb.ErrUserDuplicated
	}

	for _, p := range m.PrincipalMap {
		if reflect.DeepEqual(principal.AuthData, p.AuthData) {
			return skydb.ErrUserDuplicated
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(principal.PlainPassword), bcrypt.DefaultCost)
	if err != nil {
		panic("provider_password: Failed to hash password")
	}
	principal.HashedPassword = hashedPassword

	m.PrincipalMap[principal.ID] = principal
	return nil
}

// GetPrincipalByAuthData get principal in PrincipalMap by auth data
func (m *MockProvider) GetPrincipalByAuthData(authData map[string]interface{}, principal *Principal) error {
	for _, p := range m.PrincipalMap {
		if reflect.DeepEqual(authData, p.AuthData) {
			*principal = p
			return nil
		}
	}

	return skydb.ErrUserNotFound
}

// GetPrincipalByUserID get principal in PrincipalMap by userID
func (m *MockProvider) GetPrincipalByUserID(userID string, principal *Principal) error {
	for _, p := range m.PrincipalMap {
		if principal.UserID == userID {
			*principal = p
			return nil
		}
	}

	return skydb.ErrUserNotFound
}

// UpdatePrincipal update principal in PrincipalMap
func (m *MockProvider) UpdatePrincipal(principal Principal) error {
	if _, existed := m.PrincipalMap[principal.ID]; !existed {
		return skydb.ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(principal.PlainPassword), bcrypt.DefaultCost)
	if err != nil {
		panic("provider_password: Failed to hash password")
	}

	principal.HashedPassword = hashedPassword
	m.PrincipalMap[principal.ID] = principal
	return nil
}
