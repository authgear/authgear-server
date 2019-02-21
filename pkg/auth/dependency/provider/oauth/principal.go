package oauth

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Principal struct {
	ID              string
	UserID          string
	ProviderName    string
	ProviderUserID  string
	AccessTokenResp interface{}
	UserProfile     interface{}
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

func NewPrincipal() Principal {
	return Principal{
		ID: uuid.New(),
	}
}
