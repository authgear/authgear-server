package password

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID             string
	UserID         string
	LoginIDKey     string
	LoginID        string
	PlainPassword  string
	HashedPassword []byte
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}

func (p Principal) IsSamePassword(password string) bool {
	return bcrypt.CompareHashAndPassword(p.HashedPassword, []byte(password)) == nil
}
