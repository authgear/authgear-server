package password

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	corePassword "github.com/skygeario/skygear-server/pkg/core/password"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID             string
	UserID         string
	LoginIDKey     string
	LoginID        string
	Realm          string
	HashedPassword []byte
	ClaimsValue    map[string]interface{}
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}

func (p *Principal) setPassword(password string) (err error) {
	p.HashedPassword, err = corePassword.Hash([]byte(password))
	return
}

func (p *Principal) IsSamePassword(password string) bool {
	return corePassword.Compare([]byte(password), p.HashedPassword) == nil
}

func (p *Principal) deriveClaims(checker loginIDChecker) {
	standardKey, hasStandardKey := checker.standardKey(p.LoginIDKey)
	claimsValue := map[string]interface{}{}
	if hasStandardKey {
		claimsValue[string(standardKey)] = p.LoginID
	}
	p.ClaimsValue = claimsValue
}

func (p *Principal) PrincipalID() string {
	return p.ID
}

func (p *Principal) PrincipalUserID() string {
	return p.UserID
}

func (p *Principal) ProviderID() string {
	return string(coreAuth.PrincipalTypePassword)
}

func (p *Principal) Attributes() principal.Attributes {
	return principal.Attributes{
		"login_id_key": p.LoginIDKey,
		"login_id":     p.LoginID,
		"realm":        p.Realm,
	}
}

func (p *Principal) Claims() principal.Claims {
	return p.ClaimsValue
}
