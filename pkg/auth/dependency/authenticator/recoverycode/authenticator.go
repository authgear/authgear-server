package recoverycode

import (
	"time"
)

type Authenticator struct {
	ID        string
	UserID    string
	Code      string
	CreatedAt time.Time
	Consumed  bool
}
