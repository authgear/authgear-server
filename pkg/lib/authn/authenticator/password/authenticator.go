package password

import "time"

type Authenticator struct {
	ID           string
	Labels       map[string]interface{}
	UserID       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PasswordHash []byte
	Tag          []string
}
