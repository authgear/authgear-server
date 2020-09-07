package totp

import (
	"time"
)

type Authenticator struct {
	ID          string
	Labels      map[string]interface{}
	UserID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Secret      string
	DisplayName string
	Tag         []string
}
