package totp

import (
	"time"
)

type Authenticator struct {
	ID          string
	IsDefault   bool
	Kind        string
	UserID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Secret      string
	DisplayName string
}
