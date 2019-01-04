package password

import (
	"reflect"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"golang.org/x/crypto/bcrypt"
)

// MockProvider is the memory implementation of password provider
type MockProvider struct {
	Provider
	PrincipalMap    map[string]Principal
	authRecordKeys  [][]string
	authDataChecker authDataChecker
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(authRecordKeys [][]string) *MockProvider {
	return NewMockProviderWithPrincipalMap(authRecordKeys, map[string]Principal{})
}

// NewMockProviderWithPrincipalMap creates a new instance of mock provider with PrincipalMap
func NewMockProviderWithPrincipalMap(authRecordKeys [][]string, principalMap map[string]Principal) *MockProvider {
	return &MockProvider{
		authRecordKeys: authRecordKeys,
		authDataChecker: defaultAuthDataChecker{
			authRecordKeys: authRecordKeys,
		},
		PrincipalMap: principalMap,
	}
}

// IsAuthDataValid validates authData
func (m *MockProvider) IsAuthDataValid(authData map[string]string) bool {
	return m.authDataChecker.isValid(authData)
}

func (m *MockProvider) IsAuthDataMatching(authData map[string]string) bool {
	return m.authDataChecker.isMatching(authData)
}

// CreatePrincipalsByAuthData creates principals by authData
func (m *MockProvider) CreatePrincipalsByAuthData(authInfoID string, password string, authData map[string]string) (err error) {
	authDataList := toValidAuthDataList(m.authRecordKeys, authData)

	for _, a := range authDataList {
		principal := NewPrincipal()
		principal.UserID = authInfoID
		principal.AuthData = a
		principal.PlainPassword = password
		err = m.CreatePrincipal(principal)

		if err != nil {
			return
		}
	}

	return
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
func (m *MockProvider) GetPrincipalByAuthData(authData map[string]string, principal *Principal) (err error) {
	for _, p := range m.PrincipalMap {
		if reflect.DeepEqual(authData, p.AuthData) {
			*principal = p
			return
		}
	}

	return skydb.ErrUserNotFound
}

// GetPrincipalsByUserID get principals in PrincipalMap by userID
func (m *MockProvider) GetPrincipalsByUserID(userID string) (principals []*Principal, err error) {
	for _, p := range m.PrincipalMap {
		if p.UserID == userID {
			principal := p
			principals = append(principals, &principal)
		}
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
	}

	return
}

// GetPrincipalsByEmail get principal in PrincipalMap by userID
func (m *MockProvider) GetPrincipalsByEmail(email string) (principals []*Principal, err error) {
	for _, p := range m.PrincipalMap {
		if authData, isMap := p.AuthData.(map[string]interface{}); isMap {
			if e, found := authData["email"].(string); found && e == email {
				principal := p
				principals = append(principals, &principal)
			}
		}
	}

	if len(principals) == 0 {
		err = skydb.ErrUserNotFound
	}

	return
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
