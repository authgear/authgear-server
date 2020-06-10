package password

import (
	"time"
)

// History contains a password history of a user
type History struct {
	ID             string
	UserID         string
	HashedPassword []byte
	CreatedAt      time.Time
}
