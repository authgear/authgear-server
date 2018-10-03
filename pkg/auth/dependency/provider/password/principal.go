package password

import "github.com/skygeario/skygear-server/pkg/server/uuid"

type Principal struct {
	ID            string
	UserID        string
	AuthData      interface{}
	PlainPassword string
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}
