package password

import (
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

// MockProvider is the memory implementation of password provider
type MockProvider struct {
	Provider
	PrincipalMap   map[string]Principal
	loginIDChecker loginIDChecker
	realmChecker   realmChecker
	allowedRealms  []string
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(loginIDsKeys map[string]config.LoginIDKeyConfiguration, allowedRealms []string) *MockProvider {
	return NewMockProviderWithPrincipalMap(loginIDsKeys, allowedRealms, map[string]Principal{})
}

// NewMockProviderWithPrincipalMap creates a new instance of mock provider with PrincipalMap
func NewMockProviderWithPrincipalMap(loginIDsKeys map[string]config.LoginIDKeyConfiguration, allowedRealms []string, principalMap map[string]Principal) *MockProvider {
	return &MockProvider{
		loginIDChecker: defaultLoginIDChecker{
			loginIDsKeys: loginIDsKeys,
		},
		realmChecker: defaultRealmChecker{
			allowedRealms: allowedRealms,
		},
		allowedRealms: allowedRealms,
		PrincipalMap:  principalMap,
	}
}

// ValidateLoginIDs validates loginID
func (m *MockProvider) ValidateLoginIDs(loginIDs []LoginID) error {
	return m.loginIDChecker.validate(loginIDs)
}

func (m *MockProvider) CheckLoginIDKeyType(loginIDKey string, standardKey metadata.StandardKey) bool {
	return m.loginIDChecker.checkType(loginIDKey, standardKey)
}

func (m *MockProvider) IsRealmValid(realm string) bool {
	return m.realmChecker.isValid(realm)
}

func (m *MockProvider) IsDefaultAllowedRealms() bool {
	return len(m.allowedRealms) == 1 && m.allowedRealms[0] == DefaultRealm
}

// CreatePrincipalsByLoginID creates principals by loginID
func (m *MockProvider) CreatePrincipalsByLoginID(authInfoID string, password string, loginIDs []LoginID, realm string) (principals []*Principal, err error) {
	// do not create principal when there is login ID belongs to another user.
	for _, loginID := range loginIDs {
		loginIDPrincipals, principalErr := m.GetPrincipalsByLoginID("", loginID.Value)
		if principalErr != nil && principalErr != skydb.ErrUserNotFound {
			err = principalErr
			return
		}
		for _, principal := range loginIDPrincipals {
			if principal.UserID != authInfoID {
				err = skydb.ErrUserDuplicated
				return
			}
		}
	}

	for _, loginID := range loginIDs {
		principal := NewPrincipal()
		principal.UserID = authInfoID
		principal.LoginIDKey = loginID.Key
		principal.LoginID = loginID.Value
		principal.Realm = realm
		principal.setPassword(password)
		err = m.CreatePrincipal(principal)

		if err != nil {
			return
		}
		principals = append(principals, &principal)
	}

	return
}

// CreatePrincipal creates principal in PrincipalMap
func (m *MockProvider) CreatePrincipal(principal Principal) error {
	if _, existed := m.PrincipalMap[principal.ID]; existed {
		return skydb.ErrUserDuplicated
	}

	for _, p := range m.PrincipalMap {
		if principal.LoginID == p.LoginID && principal.Realm == p.Realm {
			return skydb.ErrUserDuplicated
		}
	}

	m.PrincipalMap[principal.ID] = principal
	return nil
}

// GetPrincipalByLoginID get principal in PrincipalMap by login_id
func (m *MockProvider) GetPrincipalByLoginIDWithRealm(loginIDKey string, loginID string, realm string, principal *Principal) (err error) {
	for _, p := range m.PrincipalMap {
		if (loginIDKey == "" || p.LoginIDKey == loginIDKey) && p.LoginID == loginID && p.Realm == realm {
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

// GetPrincipalsByLoginID get principals in PrincipalMap by login ID
func (m *MockProvider) GetPrincipalsByLoginID(loginIDKey string, loginID string) (principals []*Principal, err error) {
	for _, p := range m.PrincipalMap {
		if (loginIDKey == "" || p.LoginIDKey == loginIDKey) && p.LoginID == loginID {
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
func (m *MockProvider) UpdatePassword(principal *Principal, password string) (err error) {
	if _, existed := m.PrincipalMap[principal.ID]; !existed {
		return skydb.ErrUserNotFound
	}

	principal.setPassword(password)
	m.PrincipalMap[principal.ID] = *principal
	return nil
}
