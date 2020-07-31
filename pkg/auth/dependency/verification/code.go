package verification

import "time"

type Code struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	IdentityID string    `json:"identity_id"`
	Code       string    `json:"code"`
	ExpireAt   time.Time `json:"expire_at"`
}
