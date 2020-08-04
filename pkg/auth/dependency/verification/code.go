package verification

import "time"

type Code struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	IdentityID string `json:"identity_id"`

	LoginIDType string    `json:"login_id_type"`
	LoginID     string    `json:"login_id"`
	Code        string    `json:"code"`
	ExpireAt    time.Time `json:"expire_at"`
}
