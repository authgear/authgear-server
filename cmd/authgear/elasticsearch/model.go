package elasticsearch

import (
	"time"
)

type User struct {
	ID                string
	AppID             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastLoginAt       *time.Time
	IsDisabled        bool
	Email             []string
	PreferredUsername []string
	PhoneNumber       []string
}
