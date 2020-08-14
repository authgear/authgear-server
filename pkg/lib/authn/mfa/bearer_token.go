package mfa

import "time"

type DeviceToken struct {
	UserID    string    `json:"-"`
	Token     string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
}
