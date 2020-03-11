package userprofile

import (
	"time"
)

// Data refers the profile info of a user,
// like username, email, age, phone number
type Data map[string]interface{}

// UserProfile refers user profile data type
type UserProfile struct {
	ID        string
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time
	UpdatedBy string
	Data
}

type Store interface {
	CreateUserProfile(userID string, data Data) (UserProfile, error)
	GetUserProfile(userID string) (UserProfile, error)
	UpdateUserProfile(userID string, data Data) (UserProfile, error)
}
