package oob

import (
	"time"
)

type Code struct {
	AuthenticatorID string    `json:"authenticator_id"`
	Code            string    `json:"code"`
	ExpireAt        time.Time `json:"expire_at"`
}
