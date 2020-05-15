package bearertoken

import (
	"time"
)

type Authenticator struct {
	ID        string
	UserID    string
	ParentID  string
	Token     string
	CreatedAt time.Time
	ExpireAt  time.Time
}
