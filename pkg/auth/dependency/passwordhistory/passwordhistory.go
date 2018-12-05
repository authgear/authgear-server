package passwordhistory

import (
	"time"
)

// PasswordHistory contains a password history of a user
type PasswordHistory struct {
	ID             string
	UserID         string
	HashedPassword []byte
	LoggedAt       time.Time
}
