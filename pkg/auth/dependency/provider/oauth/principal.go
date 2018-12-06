package oauth

import (
	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

type Principal struct {
	ID             string
	UserID         string
	ProviderName   string
	ProviderUserID string
	AccessToken    interface{}
	UserProfile    interface{}
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}
