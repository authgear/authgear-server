package verification

import "time"

type Claim struct {
	ID        string
	UserID    string
	Name      string
	Value     string
	CreatedAt time.Time
}
