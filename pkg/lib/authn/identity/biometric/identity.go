package biometric

import (
	"time"
)

type Identity struct {
	ID         string
	Labels     map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UserID     string
	KeyID      string
	Key        []byte
	DeviceInfo map[string]interface{}
}
