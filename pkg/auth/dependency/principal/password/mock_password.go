package password

import (
	"strings"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

// MockProvider is the memory implementation of password provider
type MockProvider struct {
	PrincipalMap   map[string]Principal
	loginIDChecker loginIDChecker
	realmChecker   realmChecker
	allowedRealms  []string
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(loginIDsKeys []config.LoginIDKeyConfiguration, allowedRealms []string) *MockProvider {
	return NewMockProviderWithPrincipalMap(loginIDsKeys, allowedRealms, map[string]Principal{})
}

// NewMockProviderWithPrincipalMap creates a new instance of mock provider with PrincipalMap
func NewMockProviderWithPrincipalMap(loginIDsKeys []config.LoginIDKeyConfiguration, allowedRealms []string, principalMap map[string]Principal) *MockProvider {
	newFalse := func() *bool {
		t := false
		return &t
	}
	reversedNameChecker, _ := NewReservedNameChecker("../../../../../reserved_name.txt")
	return &MockProvider{
		loginIDChecker: newDefaultLoginIDChecker(
			loginIDsKeys,
			&config.LoginIDTypesConfiguration{
				Email: &config.LoginIDTypeEmailConfiguration{
					CaseSensitive: newFalse(),
					BlockPlusSign: newFalse(),
					IgnoreDotSign: newFalse(),
				},
				Username: &config.LoginIDTypeUsernameConfiguration{
					BlockReservedUsernames: newFalse(),
					ExcludedKeywords:       []string{},
					ASCIIOnly:              newFalse(),
					CaseSensitive:          newFalse(),
				},
			},
			reversedNameChecker,
		),
		realmChecker: defaultRealmChecker{
			allowedRealms: allowedRealms,
		},
		allowedRealms: allowedRealms,
		PrincipalMap:  principalMap,
	}
}

func (m *MockProvider) ValidateLoginID(loginID LoginID) error {
	return m.loginIDChecker.validateOne(loginID)
}

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

func (m *MockProvider) MakePrincipal(userID string, password string, loginID LoginID, realm string) (*Principal, error) {
	principal := NewPrincipal()
	principal.UserID = userID
	principal.LoginIDKey = loginID.Key
	principal.LoginID = loginID.Value
	principal.Realm = realm
	principal.setPassword(password)
	principal.deriveClaims(m.loginIDChecker)
	return &principal, nil
}

// CreatePrincipalsByLoginID creates principals by loginID
func (m *MockProvider) CreatePrincipalsByLoginID(userID string, password string, loginIDs []LoginID, realm string) (principals []*Principal, err error) {
	// do not create principal when there is login ID belongs to another user.
	for _, loginID := range loginIDs {
		loginIDPrincipals, principalErr := m.GetPrincipalsByLoginID("", loginID.Value)
		if principalErr != nil && principalErr != principal.ErrNotFound {
			err = principalErr
			return
		}
		for _, p := range loginIDPrincipals {
			if p.UserID != userID {
				err = ErrLoginIDAlreadyUsed
				return
			}
		}
	}

	for _, loginID := range loginIDs {
		principal, _ := m.MakePrincipal(userID, password, loginID, realm)
		err = m.CreatePrincipal(principal)

		if err != nil {
			return
		}
		principals = append(principals, principal)
	}

	return
}

// CreatePrincipal creates principal in PrincipalMap
func (m *MockProvider) CreatePrincipal(p *Principal) error {
	if _, existed := m.PrincipalMap[p.ID]; existed {
		return principal.ErrAlreadyExists
	}

	for _, pp := range m.PrincipalMap {
		if p.LoginID == pp.LoginID && p.Realm == pp.Realm {
			return principal.ErrAlreadyExists
		}
	}

	m.PrincipalMap[p.ID] = *p
	return nil
}

// GetPrincipalByLoginID get principal in PrincipalMap by login_id
func (m *MockProvider) GetPrincipalByLoginIDWithRealm(loginIDKey string, loginID string, realm string, p *Principal) (err error) {
	for _, pp := range m.PrincipalMap {
		if (loginIDKey == "" || pp.LoginIDKey == loginIDKey) && strings.EqualFold(pp.LoginID, loginID) && pp.Realm == realm {
			*p = pp
			return
		}
	}

	return principal.ErrNotFound
}

// GetPrincipalsByUserID get principals in PrincipalMap by userID
func (m *MockProvider) GetPrincipalsByUserID(userID string) (principals []*Principal, err error) {
	for _, p := range m.PrincipalMap {
		if p.UserID == userID {
			principal := p
			principals = append(principals, &principal)
		}
	}

	return
}

// GetPrincipalsByLoginID get principals in PrincipalMap by login ID
func (m *MockProvider) GetPrincipalsByLoginID(loginIDKey string, loginID string) (principals []*Principal, err error) {
	for _, p := range m.PrincipalMap {
		if (loginIDKey == "" || p.LoginIDKey == loginIDKey) && strings.EqualFold(p.LoginID, loginID) {
			principal := p
			principals = append(principals, &principal)
		}
	}

	return
}

func (m *MockProvider) UpdatePassword(p *Principal, password string) (err error) {
	if _, existed := m.PrincipalMap[p.ID]; !existed {
		return principal.ErrNotFound
	}

	p.setPassword(password)
	m.PrincipalMap[p.ID] = *p
	return nil
}

func (m *MockProvider) MigratePassword(p *Principal, password string) (err error) {
	if _, existed := m.PrincipalMap[p.ID]; !existed {
		return principal.ErrNotFound
	}

	p.migratePassword(password)
	m.PrincipalMap[p.ID] = *p
	return nil
}

func (m *MockProvider) ID() string {
	return string(coreAuth.PrincipalTypePassword)
}

func (m *MockProvider) GetPrincipalByID(principalID string) (principal.Principal, error) {
	for _, p := range m.PrincipalMap {
		if p.ID == principalID {
			return &p, nil
		}
	}
	return nil, principal.ErrNotFound
}

func (m *MockProvider) ListPrincipalsByClaim(claimName string, claimValue string) ([]principal.Principal, error) {
	var principals []principal.Principal
	for _, p := range m.PrincipalMap {
		if p.ClaimsValue[claimName] == claimValue {
			principal := p
			principals = append(principals, &principal)
		}
	}
	return principals, nil
}

func (m *MockProvider) ListPrincipalsByUserID(userID string) ([]principal.Principal, error) {
	var principals []principal.Principal
	for _, p := range m.PrincipalMap {
		if p.UserID == userID {
			principal := p
			principals = append(principals, &principal)
		}
	}
	return principals, nil
}

var (
	_ Provider = &MockProvider{}
)
