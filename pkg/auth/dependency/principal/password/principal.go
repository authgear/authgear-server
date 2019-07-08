package password

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"golang.org/x/crypto/bcrypt"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID             string
	UserID         string
	LoginIDKey     string
	LoginID        string
	Realm          string
	HashedPassword []byte
}

type attributes struct {
	LoginIDKey string `json:"login_id_key"`
	LoginID    string `json:"login_id"`
	Realm      string `json:"realm"`
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}

func (p *Principal) setPassword(password string) (err error) {
	p.HashedPassword, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return
}

func (p *Principal) IsSamePassword(password string) bool {
	return bcrypt.CompareHashAndPassword(p.HashedPassword, []byte(password)) == nil
}

func (p *Principal) PrincipalID() string {
	return p.ID
}

func (p *Principal) PrincipalUserID() string {
	return p.UserID
}

func (p *Principal) ProviderType() string {
	return providerPassword
}

func (p *Principal) Attributes() principal.Attributes {
	return attributes{
		LoginIDKey: p.LoginIDKey,
		LoginID:    p.LoginID,
		Realm:      p.Realm,
	}
}
