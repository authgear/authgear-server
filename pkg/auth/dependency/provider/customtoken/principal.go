package customtoken

import (
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type Principal struct {
	ID               string
	TokenPrincipalID string
	UserID           string
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}
