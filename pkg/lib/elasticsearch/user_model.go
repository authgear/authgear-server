package elasticsearch

import (
	"time"
)

const IndexNameUser = "user"

type User struct {
	ID                string     `json:"id,omitempty"`
	AppID             string     `json:"app_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at,omitempty"`
	UpdatedAt         time.Time  `json:"updated_at,omitempty"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	IsDisabled        bool       `json:"is_disabled"`
	Email             []string   `json:"email,omitempty"`
	PreferredUsername []string   `json:"preferred_username,omitempty"`
	PhoneNumber       []string   `json:"phone_number,omitempty"`
}
