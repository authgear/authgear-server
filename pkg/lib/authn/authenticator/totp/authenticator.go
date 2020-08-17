package totp

import (
	"time"
)

type Authenticator struct {
	ID          string
	UserID      string
	CreatedAt   time.Time
	Secret      string
	DisplayName string
	Tag         []string
}
