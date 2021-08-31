package mfa

import (
	"time"
)

type RecoveryCode struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Consumed  bool      `json:"consumed"`
}
