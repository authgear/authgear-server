package password

import (
	"reflect"

	"github.com/skygeario/skygear-server/pkg/core/skydb"
	"golang.org/x/crypto/bcrypt"
)

// MockProvider is the memory implementation of password provider
type MockProvider struct {
	Provider
	PrincipalMap         map[string]Principal
	loginIDsKeyWhitelist []string
	loginIDChecker       loginIDChecker
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(loginIDsKeyWhitelist []string) *MockProvider {
	return NewMockProviderWithPrincipalMap(loginIDsKeyWhitelist, map[string]Principal{})
}

// NewMockProviderWithPrincipalMap creates a new instance of mock provider with PrincipalMap
func NewMockProviderWithPrincipalMap(loginIDsKeyWhitelist []string, principalMap map[string]Principal) *MockProvider {
	return &MockProvider{
		loginIDsKeyWhitelist: loginIDsKeyWhitelist,
		loginIDChecker: defaultLoginIDChecker{
			loginIDsKeyWhitelist: loginIDsKeyWhitelist,
		},
		PrincipalMap: principalMap,
	}
}

// IsLoginIDValid validates loginID
func (m *MockProvider) IsLoginIDValid(loginID map[string]string) bool {
	return m.loginIDChecker.isValid(loginID)
}

func (m *MockProvider) IsLoginIDMatching(loginID map[string]string) bool {
	return m.loginIDChecker.isMatching(loginID)
}

// CreatePrincipalsByLoginID creates principals by loginID
func (m *MockProvider) CreatePrincipalsByLoginID(authInfoID string, password string, loginID map[string]string) (err error) {
	loginIDList := toValidLoginIDMap(m.loginIDsKeyWhitelist, loginID)

	for k, v := range loginIDList {
		principal := NewPrincipal()
		principal.UserID = authInfoID
		principal.LoginIDKey = k
		principal.LoginID = v
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
		if reflect.DeepEqual(principal.LoginID, p.LoginID) {
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

// GetPrincipalByLoginID get principal in PrincipalMap by login_id
func (m *MockProvider) GetPrincipalByLoginID(loginIDKey string, loginID string, principal *Principal) (err error) {
	for _, p := range m.PrincipalMap {
		if p.LoginIDKey == loginIDKey && p.LoginID == loginID {
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
		if p.LoginIDKey == "email" && p.LoginID == email {
			principal := p
			principals = append(principals, &principal)
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
