package siwe

import "time"

type Nonce struct {
	Nonce    string    `json:"nonce"`
	ExpireAt time.Time `json:"expire_at"`
}
